// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fakestorage

import (
	"crypto/md5" // #nosec G501
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	// "mime"
	// "mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"encoding/base64"
	"hash/crc32"

	"github.com/gorilla/mux"
)

type multipartMetadata struct {
	Name string `json:"name"`
}

func (s *Server) insertObject(w http.ResponseWriter, r *http.Request) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	bucketName := mux.Vars(r)["bucketName"]
	if err := s.backend.GetBucket(bucketName); err != nil {
		w.WriteHeader(http.StatusNotFound)
		err := newErrorResponse(http.StatusNotFound, "Not found", nil)
		json.NewEncoder(w).Encode(err)
		return
	}
	uploadType := r.URL.Query().Get("uploadType")
	fmt.Println(fmt.Sprintf("Inserting object '%s'", uploadType))

	fmt.Println(formatRequest(r))
	switch uploadType {
	case "media":
		s.simpleUpload(bucketName, w, r)
	case "multipart":
		s.multipartUpload(bucketName, w, r)
	case "resumable":
		s.resumableUpload(bucketName, w, r)
	default:
		http.Error(w, "invalid uploadType", http.StatusBadRequest)
	}
}

func (s *Server) simpleUpload(bucketName string, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name is required for simple uploads", http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	obj := Object{BucketName: bucketName, Name: name, Content: data, Crc32c: encodedCrc32cChecksum(data), Md5Hash: encodedMd5Hash(data)}
	err = s.createObject(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(obj)
}

var crc32cTable = crc32.MakeTable(crc32.Castagnoli)

func crc32cChecksum(content []byte) []byte {
	checksummer := crc32.New(crc32cTable)
	checksummer.Write(content)
	return checksummer.Sum(make([]byte, 0, 4))
}

func encodedChecksum(checksum []byte) string {
	return base64.StdEncoding.EncodeToString(checksum)
}

func encodedCrc32cChecksum(content []byte) string {
	return encodedChecksum(crc32cChecksum(content))
}

func md5Hash(b []byte) []byte {
	/* #nosec G401 */
	h := md5.New()
	h.Write(b)
	return h.Sum(nil)
}

func encodedHash(hash []byte) string {
	return base64.StdEncoding.EncodeToString(hash)
}

func encodedMd5Hash(content []byte) string {
	return encodedHash(md5Hash(content))
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
	  name = strings.ToLower(name)
	  for _, h := range headers {
		request = append(request, fmt.Sprintf("%v: %v\n", name, h))
	  }
	}
	
	r.ParseForm()

	r.ParseMultipartForm(1234567)

	fmt.Println("Form data")

	for k, v := range r.Form {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
	}
	
	fmt.Println("PostForm data")

	for k, v := range r.PostForm {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
	}
	
	// fmt.Println("MultipartForm data")

	// for k, v := range r.MultipartForm {
    //     fmt.Println("key:", k)
    //     fmt.Println("val:", strings.Join(v, ""))
    // }
	
	// If this is a POST, add post data
	if r.Method == "POST" {
	   r.ParseForm()
	   request = append(request, "\n")
	   request = append(request, r.Form.Encode())
	} 
	 // Return the request as a string
	 return strings.Join(request, "\n")
   }


func (s *Server) multipartUpload(bucketName string, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// bod, _ := ioutil.ReadAll(r.Body)

	// st := string(bod)

	// fmt.Println(st)

	// contentType := r.Header.Get("Content-Type")





	// mediaType, params, err := mime.ParseMediaType(contentType)
	// if err != nil {
	//   panic(err)
	// }
	// if mediaType != "multipart/related" {
	//   panic(nil)
	// }

	// for k, _ := range r.MultipartForm.File {
	// 	fmt.Println(k)
	// }

	// rm := NewReader(r.Body, params)
	// object, err := rm.ReadObject()
	// if err != nil {
	//   panic(err)
	// }
	// for i, part := range object.Values {
	//   b, err := ioutil.ReadAll(part)
	//   if err != nil {
	// 	panic(err)
	//   }
	//   fmt.Printf("%d.: %s \n", i, string(b))
	// }







	// _, params, err := mime.ParseMediaType(contentType)
	// if err != nil {
	// 	http.Error(w, "invalid Content-Type header", http.StatusBadRequest)
	// 	return
	// }
	var (
		// metadata *multipartMetadata
		content  []byte
	)

	// reader := multipart.NewReader(r.Body, params["boundary"])
	// fmt.Println(fmt.Sprintf("boundary %s", params["boundary"]))
	// part, err := reader.NextPart()
	// for ; err == nil; part, err = reader.NextPart() {
	// 	if metadata == nil {
	// 		metadata, err = loadMetadata(part)
	// 	} else {
	// 		content, err = loadContent(part)
	// 	}
	// 	if err != nil {
	// 		break
	// 	}
	// }
	// if err != io.EOF {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// obj := Object{BucketName: bucketName, Name: metadata.Name, Content: content, Crc32c: encodedCrc32cChecksum(content), Md5Hash: encodedMd5Hash(content)}
	
	content = []byte("{}")

	fmt.Println(fmt.Sprintf("Bucket name=%s", bucketName))
	obj := Object{BucketName: bucketName, Bucket: bucketName, Name: bucketName, Content: content, Crc32c: encodedCrc32cChecksum(content), Md5Hash: encodedMd5Hash(content)}
	// err = s.createObject(obj)
	err := s.createObject(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(obj)
}

func (s *Server) resumableUpload(bucketName string, w http.ResponseWriter, r *http.Request) {
	objName := r.URL.Query().Get("name")
	if objName == "" {
		metadata, err := loadMetadata(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		objName = metadata.Name
	}
	obj := Object{BucketName: bucketName, Name: objName}
	uploadID, err := generateUploadID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.uploads[uploadID] = obj
	w.Header().Set("Location", s.URL()+"/upload/resumable/"+uploadID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(obj)
}

func (s *Server) uploadFileContent(w http.ResponseWriter, r *http.Request) {
	uploadID := mux.Vars(r)["uploadId"]
	s.mtx.Lock()
	defer s.mtx.Unlock()
	obj, ok := s.uploads[uploadID]
	if !ok {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}
	content, err := loadContent(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	commit := true
	status := http.StatusCreated
	objLength := len(obj.Content)
	obj.Content = append(obj.Content, content...)
	obj.Crc32c = encodedCrc32cChecksum(obj.Content)
	obj.Md5Hash = encodedMd5Hash(obj.Content)
	if contentRange := r.Header.Get("Content-Range"); contentRange != "" {
		commit, err = parseRange(contentRange, objLength, len(content), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if commit {
		delete(s.uploads, uploadID)
		err = s.createObject(obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		status = http.StatusOK
		w.Header().Set("X-Http-Status-Code-Override", "308")
		s.uploads[uploadID] = obj
	}
	data, _ := json.Marshal(obj)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	w.Write(data)
}

func parseRange(r string, objLength, bodyLength int, w http.ResponseWriter) (finished bool, err error) {
	invalidErr := fmt.Errorf("invalid Content-Range: %v", r)
	const bytesPrefix = "bytes "
	var contentLength int
	if !strings.HasPrefix(r, bytesPrefix) {
		return false, invalidErr
	}
	parts := strings.SplitN(r[len(bytesPrefix):], "/", 2)
	if len(parts) != 2 {
		return false, invalidErr
	}
	var rangeStart, rangeEnd int

	if parts[0] == "*" {
		rangeStart = objLength
		rangeEnd = objLength + bodyLength
	} else {
		rangeParts := strings.SplitN(parts[0], "-", 2)
		if len(rangeParts) != 2 {
			return false, invalidErr
		}
		rangeStart, err = strconv.Atoi(rangeParts[0])
		if err != nil {
			return false, invalidErr
		}
		rangeEnd, err = strconv.Atoi(rangeParts[1])
		if err != nil {
			return false, invalidErr
		}
	}

	contentLength = objLength + bodyLength
	finished = rangeEnd == contentLength
	w.Header().Set("Range", fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd))

	return finished, nil
}

func loadMetadata(rc io.ReadCloser) (*multipartMetadata, error) {
	fmt.Println("Loading metadata")
	defer rc.Close()
	var m multipartMetadata
	err := json.NewDecoder(rc).Decode(&m)
	return &m, err
}

func loadContent(rc io.ReadCloser) ([]byte, error) {
	fmt.Println("Loading content")
	defer rc.Close()
	return ioutil.ReadAll(rc)
}

func generateUploadID() (string, error) {
	var raw [16]byte
	_, err := rand.Read(raw[:])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", raw[:]), nil
}
