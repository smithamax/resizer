package resizer

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
)

type handler struct {
	source Source
}

func Handler(source Source) http.Handler {
	return &handler{source}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err *appError

	switch r.Method {
	case http.MethodGet:
		err = h.handleGet(w, r)
	case http.MethodPost:
		err = h.handlePost(w, r)
	default:
		http.Error(w, "Only GET and POST supported", http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		log.Printf("%s %s %d : %v", r.Method, r.RequestURI, err.Code, err.Error)
		http.Error(w, err.Message, err.Code)
	}
}

type appError struct {
	Error   error
	Message string
	Code    int
}

func (h *handler) handleGet(w http.ResponseWriter, r *http.Request) *appError {
	q := r.URL.Query()
	path := r.URL.Path
	width, _ := strconv.ParseInt(q.Get("w"), 10, 0)
	height, _ := strconv.ParseInt(q.Get("h"), 10, 0)
	raw, _ := strconv.ParseBool(q.Get("raw"))
	fit := q.Get("fit")

	fmt.Println(path)

	imgr, err := h.source.Get(path)
	if err != nil {
		return &appError{err, "Error getting image", http.StatusInternalServerError}
	}
	if imgr == nil {
		return &appError{err, "Image not found", http.StatusNotFound}
	}
	defer imgr.Close()
	if raw {
		// TODO: Add format header for raw images
		// w.Header().Set("Content-Type", fmt.Sprintf("image/%s", format))
		if _, err = io.Copy(w, imgr); err != nil {
			return &appError{err, "Error writing response", http.StatusInternalServerError}
		}
	}

	img, format, err := NormaliseDecode(imgr)
	if err != nil {
		return &appError{err, "Error decoding image", http.StatusInternalServerError}
	}

	img = Resize(img, int(width), int(height), fit)

	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", format))
	if err = Encode(w, img, format); err != nil {
		return &appError{err, "Error writing response", http.StatusInternalServerError}
	}
	return nil
}

type uploadResponse struct {
	Path         string `json:"path"`
	PathTemplate string `json:"pathTemplate"`
}

func (h *handler) handlePost(w http.ResponseWriter, r *http.Request) *appError {
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
			return &appError{err, "Error reading form file", http.StatusBadRequest}
		}
	} else {
		file = r.Body
	}
	io.Copy(buf, io.TeeReader(file, hash))
	file.Close()

	name := fmt.Sprintf("%x", hash.Sum(nil))

	img, format, err := NormaliseDecode(buf)
	if err != nil {
		return &appError{err, "Error decoding image", http.StatusBadRequest}
	}

	err = Encode(buf, img, format)
	if err != nil {
		return &appError{err, "Error re-encoding image", http.StatusInternalServerError}
	}

	err = h.source.Put(name, buf, "image/"+format)
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
