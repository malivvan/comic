package comic

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestComic(t *testing.T) {
	path := "book_test_file.cbz"
	defer os.Remove(path)
	defer os.Remove(path + ".tmp")

	// create test image file
	bigBuff := make([]byte, 1000000)
	err := ioutil.WriteFile("sample.jpg", bigBuff, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("sample.jpg")

	// create book
	book, err := Create(path, "The Walking Dead", "tset", "oOoO")
	if err != nil {
		t.Fatal(err)
	}
	if book.Title != "The Walking Dead" || book.Artist != "tset" || book.Language != "oOoO" {
		t.Fatal("meta creation failed")
	}
	if book.fileprefix() != "the_walking_dead_" {
		t.Fatal("bad fileprefix")
	}
	err = book.Add([]string{"sample.jpg"})
	if err != nil {
		t.Fatal(err)
	}

	// reopen book
	book2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if book2.Title != "The Walking Dead" || book2.Artist != "tset" || book2.Language != "oOoO" {
		t.Fatal("reloading meta failed")
	}
	err = book2.Add([]string{"sample.jpg"})
	if err != nil {
		t.Fatal(err)
	}

	// check pages
	for i := 0; i < book2.Pages(); i++ {
		info := book2.GetPageName(i)
		if i == 0 && info != "the_walking_dead_0001.jpg" {
			t.Fatal("error generating filename")
		}
		if i == 1 && info != "the_walking_dead_0002.jpg" {
			t.Fatal("error generating filename")
		}
	}
}
