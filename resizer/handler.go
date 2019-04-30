package resizer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
)

func Handler(source Source) http.Handler {
	get := handlerMetric(appErrorHandler(getHandler(source)), "fetch")
	post := handlerMetric(appErrorHandler(postHandler(source)), "upload")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			get.ServeHTTP(w, r)
		case http.MethodPost:
			post.ServeHTTP(w, r)
		default:
			http.Error(w, "Only GET and POST supported", http.StatusMethodNotAllowed)
			return
		}
	})
}

type appError struct {
	Error   error
	Message string
	Code    int
}

type appHandler func(w http.ResponseWriter, r *http.Request) *appError

func appErrorHandler(h appHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			log.Printf("%s %s %d : %v", r.Method, r.RequestURI, err.Code, err.Error)
			http.Error(w, err.Message, err.Code)
		}
	})
}

func getHandler(source Source) appHandler {
	return func(w http.ResponseWriter, r *http.Request) *appError {
		q := r.URL.Query()
		path := r.URL.Path
		width, _ := strconv.ParseInt(q.Get("w"), 10, 0)
		height, _ := strconv.ParseInt(q.Get("h"), 10, 0)
		raw, _ := strconv.ParseBool(q.Get("raw"))
		fit := q.Get("fit")

		imgr, err := source.Get(path)
		if err != nil {
			return &appError{err, "Error getting image", http.StatusInternalServerError}
		}
		if imgr == nil {
			// For now just cache 404s for a little while
			w.Header().Set("Cache-Control", "public, max-age=180")
			return &appError{err, "Image not found", http.StatusNotFound}
		}
		defer imgr.Close()
		if raw {
			// TODO: Add format header for raw images
			// w.Header().Set("Content-Type", fmt.Sprintf("image/%s", format))
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			if _, err = io.Copy(w, imgr); err != nil {
				return &appError{err, "Error writing response", http.StatusInternalServerError}
			}
		}

		b, err := ioutil.ReadAll(imgr)
		if err != nil {
			return &appError{err, "Error reading image", http.StatusInternalServerError}
		}

		img, format, err := Transform(b, int(width), int(height), fit)
		if err != nil {
			return &appError{err, "Error transforming image", http.StatusInternalServerError}
		}
		compressionRatio.WithLabelValues().Observe(float64(len(img)) / float64(len(b)))

		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Header().Set("Content-Type", fmt.Sprintf("image/%s", format))
		if _, err = w.Write(img); err != nil {
			return &appError{err, "Error writing response", http.StatusInternalServerError}
		}
		return nil
	}
}

type uploadResponse struct {
	Path         string `json:"path"`
	PathTemplate string `json:"pathTemplate"`
}

func postHandler(source Source) appHandler {
	return func(w http.ResponseWriter, r *http.Request) *appError {
		hash := sha256.New()
		buf := &bytes.Buffer{}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			panic(err)
		}

		var file io.ReadCloser
		if mediaType == "multipart/form-data" {
			file, _, err = r.FormFile("file")
			if err != nil {
				return &appError{err, "Error reading form", http.StatusBadRequest}
			}
		} else {
			file = r.Body
		}
		b, err := ioutil.ReadAll(io.TeeReader(file, hash))
		if err != nil {
			return &appError{err, "Error reading file", http.StatusBadRequest}
		}
		file.Close()

		name := path.Join(r.URL.Path, hex.EncodeToString(hash.Sum(nil)))

		img, format, err := NormaliseDecode(bytes.NewReader(b))
		if err != nil {
			return &appError{err, "Error decoding image", http.StatusBadRequest}
		}

		err = Encode(buf, img, format)
		if err != nil {
			return &appError{err, "Error re-encoding image", http.StatusInternalServerError}
		}

		err = source.Put(name, buf, "image/"+format)
		if err != nil {
			return &appError{err, "Error persisting image", http.StatusInternalServerError}
		}

		res, err := json.Marshal(uploadResponse{
			Path:         fmt.Sprintf("images/%s", name),
			PathTemplate: fmt.Sprintf("images/%s{?w,h,raw}", name),
		})

		if err != nil {
			return &appError{err, "Error encoding json", http.StatusInternalServerError}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
		return nil
	}
}
