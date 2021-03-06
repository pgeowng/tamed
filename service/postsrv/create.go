package postsrv

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/pgeowng/tamed/model"
	"github.com/pgeowng/tamed/service/commonsrv"
	"github.com/pgeowng/tamed/types"
	"github.com/pkg/errors"
)

func (srv *PostSrv) Create(files []*multipart.FileHeader) ([]model.PostCreate, error) {
	fmt.Printf("postsrv.create")

	for idx, header := range files {
		if header.Size == 0 {
			return nil, errors.Errorf("postsrv.create: empty data inside %d file", idx)
		}
	}

	result := make([]model.PostCreate, 0, len(files))
	for _, header := range files {
		result = append(result, srv.CreateFile(header))
	}

	return result, nil
}

func GuessContentType(file io.ReadCloser) (string, error) {
	defer file.Close()

	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return "", errors.Wrap(err, "guessmime")
	}
	contentType := http.DetectContentType(buf)
	return contentType, nil
}
func (srv *PostSrv) CreateFile(header *multipart.FileHeader) model.PostCreate {
	fileName := header.Filename
	file, err := header.Open()
	if err != nil {
		return model.PostCreate{
			Error: err.Error(),
		}
	}

	contentType, err := GuessContentType(file)
	fmt.Println("Content Type: " + contentType)
	file.Close()

	_, ok := types.AcceptedMime[contentType]
	if !ok {
		err := errors.Errorf("postsrv.create(%s): bad upload type %v", fileName, contentType)
		return model.PostCreate{
			Error: err.Error(),
		}
	}

	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")

	if len(ext) == 0 {
		myExt := types.GetExt(contentType)
		if len(myExt) == 0 {
			err := errors.Errorf("postsrv.create(%s): cant detect content type", fileName)
			return model.PostCreate{
				Error: err.Error(),
			}
		}
		ext = myExt
	}

	id := commonsrv.UniqID()

	file, err = header.Open()
	if err != nil {
		err = errors.Wrap(err, "postsrv.create("+fileName+")")
		return model.PostCreate{
			Error: err.Error(),
		}
	}

	filePath, err := srv.store.Media.Upload(id, ext, file)
	file.Close()
	if err != nil {
		err = errors.Wrap(err, "postsrv.create("+fileName+")")
		return model.PostCreate{
			Error: err.Error(),
		}
	}

	obj := model.Post{
		PostID:     id,
		CreateTime: commonsrv.TimeNow(),
		Tags:       model.NewTags(),
		Link:       filePath,
	}

	err = srv.store.Post.Create(id, &obj)
	if err != nil {
		err = errors.Wrap(err, "postsrv.create("+fileName+")")
		return model.PostCreate{
			Error: err.Error(),
		}
	}

	return model.PostCreate{
		PostID:     obj.PostID,
		CreateTime: obj.CreateTime,
		Tags:       obj.Tags,
		Link:       obj.Link,
	}
}
