package shelf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const EntityPath = "entity"

type entityShelf struct {
	storage  *Storage
	entities map[string]Entity
}

func newEntityShelf(storage *Storage) *entityShelf {
	return &entityShelf{
		storage:  storage,
		entities: make(map[string]Entity),
	}
}

func (s *entityShelf) Size() int {
	return len(s.entities)
}

func loadMetadata(storage *Storage, path string, out Entity) (Entity, error) {
	log := zap.L()
	f, err := storage.OpenFile(path)
	if err != nil {
		log.Fatal("Failed to open file", zap.String("path", path), zap.Error(err))
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *entityShelf) Init() error {
	log := zap.L()
	return s.storage.WalkFiles(EntityPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var e Entity
		dirName, fileName := filepath.Split(path)
		if strings.HasSuffix(fileName, ".image.yml") {
			ent := &ImageEntity{}
			ent.ID_ = strings.TrimSuffix(fileName, ".image.yml")
			ent.MetaPath_ = path
			if e, err = loadMetadata(s.storage, path, ent); err != nil {
				return err
			}
			switch ent.MimeType_ {
			case "image/gif":
				ent.Path_ = filepath.Join(dirName, ent.ID_) + ".gif"
			case "image/jpeg":
				ent.Path_ = filepath.Join(dirName, ent.ID_) + ".jpg"
			case "image/png":
				ent.Path_ = filepath.Join(dirName, ent.ID_) + ".png"
			default:
				log.Fatal("Unknown image type", zap.String("mime-type", ent.MimeType_))
			}
			ent.SystemPath_ = s.storage.path(ent.Path_)
		} else if strings.HasSuffix(fileName, ".video.yml") {
			ent := &VideoEntity{}
			ent.ID_ = strings.TrimSuffix(fileName, ".video.yml")
			ent.MetaPath_ = path
			if e, err = loadMetadata(s.storage, path, ent); err != nil {
				return err
			}
			switch ent.MimeType_ {
			case "video/mp4":
				ent.Path_ = filepath.Join(dirName, ent.ID_) + ".mp4"
			default:
				log.Fatal("Unknown video type", zap.String("mime-type", ent.MimeType_))
			}
			ent.SystemPath_ = s.storage.path(ent.Path_)
		} else if strings.HasSuffix(fileName, ".audio.yml") {
			ent := &AudioEntity{}
			ent.ID_ = strings.TrimSuffix(fileName, ".audio.yml")
			ent.MetaPath_ = path
			if e, err = loadMetadata(s.storage, path, ent); err != nil {
				return err
			}
			switch ent.MimeType_ {
			default:
				log.Fatal("Unknwon audio type", zap.String("mime-type", ent.MimeType_))
			}
			ent.Path_ = s.storage.path(ent.Path_)
		} else {
			//Continue...
			return nil
		}
		if !s.storage.Exists(e.Path()) {
			return fmt.Errorf("file not found: %s", e.Path())
		}
		expectedDir := s.calcDir(e)
		if filepath.Dir(e.Path()) != expectedDir {
			log.Warn("Dir mismatched", zap.String("expected", expectedDir), zap.String("actual", filepath.Dir(e.Path())))
			// TODO: move?
		}
		s.entities[e.ID()] = e
		log.Debug("Reading Entity", zap.String("id", e.ID()), zap.String("mime-type", e.MimeType()))
		return nil
	})
}

func (s *entityShelf) Remove(e Entity) error {
	switch entity := e.(type) {
	case *ImageEntity:
		return s.RemoveImage(entity)
	case *VideoEntity:
		return s.RemoveVideo(entity)
	default:
		return fmt.Errorf("unkown entity: %v", reflect.TypeOf(e))
	}
}

func (s *entityShelf) calcDir(e Entity) string {
	return filepath.Join(EntityPath, strconv.Itoa(e.Date().Year()))
}

func (s *entityShelf) Lookup(id string) Entity {
	return s.entities[id]
}

func (s *entityShelf) AsSlice() []Entity {
	i := 0
	lst := make([]Entity, len(s.entities))
	for _, e := range s.entities {
		lst[i] = e
		i++
	}
	return lst
}
