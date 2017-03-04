package comic

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

var imageFileSuffix = []string{".png", ".jpeg", ".jpg"}

type Book struct {
	Title    string
	Artist   string
	Language string
	path     string
	pages    []string
}

func (book *Book) GetPageName(index int) string {
	if index < len(book.pages) {
		return book.pages[index]
	}
	return ""
}

func Open(path string) (*Book, error) {
	var book Book
	book.path = path
	err := book.load()
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func Create(path string, title string, artist string, language string) (*Book, error) {
	book := &Book{
		Title:    title,
		Artist:   artist,
		Language: language,
		path:     path,
		pages:    []string{},
	}
	err := book.save(nil, nil)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (book *Book) Add(sources []string) error {
	err := book.save(sources, nil)
	if err != nil {
		return err
	}
	err = book.load()
	if err != nil {
		return err
	}
	return nil
}

func (book *Book) Remove(pages []int) error {
	return book.save(nil, pages)
}

func (book *Book) Pages() int {
	return len(book.pages)
}

func (book *Book) GetPage(index int) (io.ReadCloser, error) {
	if !(index < book.Pages()) {
		return nil, errors.New("page does not exist")
	}
	name := book.pages[index]
	reader, err := zip.OpenReader(book.path)
	if err != nil {
		return nil, err
	}
	for _, file := range reader.File {
		if file.Name == name {
			return file.Open()
		}
	}
	return nil, errors.New("page does not exist")
}

func (book *Book) fileprefix() string {
	if book.Title == "" {
		return ""
	}
	return strings.Replace(strings.ToLower(book.Title), " ", "_", -1) + "_"
}

func (book *Book) nextFilename(s string) (string, error) {

	// get file suffix
	fileSuffix := ""
	for _, suffix := range imageFileSuffix {
		if strings.HasSuffix(s, suffix) {
			fileSuffix = suffix
		}
	}
	if fileSuffix == "" {
		return "", errors.New("not a valid image file")
	}

	// first file
	if len(book.pages) == 0 {
		return book.fileprefix() + "0001" + fileSuffix, nil
	}

	// from last pages name
	digit := 0
	fileName := []rune(strings.TrimSuffix(book.pages[len(book.pages)-1], fileSuffix))
	for i := len(fileName) - 1; i > 0; i-- {
		r := string(fileName[i])
		if strings.Contains(r, "9") {
			digit++
			continue
		}
		if strings.ContainsAny(r, "012345678") {
			ri, err := strconv.Atoi(r)
			if err != nil {
				return "", err
			}
			ri++
			fileName[i] = rune(strconv.Itoa(ri)[0])
			for n := 1; n <= digit; n++ {
				fileName[i+n] = '0'
			}
			break
		}
		return "", errors.New("no page numbers left ")

	}
	return string(fileName) + fileSuffix, nil
}

func (book *Book) save(sources []string, remove []int) error {
	err := book.saveTmp(sources, remove)
	if err != nil {
		return err
	}
	return os.Rename(book.path+".tmp", book.path)
}

func (book *Book) saveTmp(sources []string, remove []int) error {
	tmpPath := book.path + ".tmp"
	os.Remove(tmpPath)

	// create tmpfile and archive writer
	zipfile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer zipfile.Close()
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// write meta
	meta, err := json.Marshal(book)
	if err != nil {
		return err
	}
	metaWriter, err := archive.Create("meta.json")
	if err != nil {
		return err
	}
	_, err = metaWriter.Write(meta)
	if err != nil {
		return err
	}

	// copy over files from exsisting archive
	if info, err := os.Stat(book.path); err == nil && !info.IsDir() {
		for i := 0; i < book.Pages(); i++ {
			r, err := book.GetPage(i)
			if err != nil {
				return err
			}
			w, err := archive.Create(book.pages[i])
			if err != nil {
				return err
			}
			_, err = io.Copy(w, r)
			if err != nil {
				return err
			}
			err = r.Close()
			if err != nil {
				return err
			}
		}
	}

	// append new sources
	if sources != nil {
		for _, source := range sources {
			r, err := os.Open(source)
			if err != nil {
				return err
			}
			name, err := book.nextFilename(source)
			if err != nil {
				return err
			}
			book.pages = append(book.pages, name)
			w, err := archive.Create(name)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, r)
			if err != nil {
				return err
			}
			err = r.Close()
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (book *Book) load() error {
	book.pages = []string{}
	reader, err := zip.OpenReader(book.path)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		info := file.FileInfo()
		if info.Name() == "meta.json" {
			meta, err := file.Open()
			if err != nil {
				return err
			}
			d := json.NewDecoder(meta)
			err = d.Decode(&book)
			if err != nil {
				return err
			}
			err = meta.Close()
			if err != nil {

				return err
			}
		} else {
			for _, suffix := range imageFileSuffix {
				if strings.HasSuffix(file.Name, suffix) {
					book.pages = append(book.pages, file.Name)
				}
			}
		}
	}
	sort.Strings(book.pages)
	return nil
}
