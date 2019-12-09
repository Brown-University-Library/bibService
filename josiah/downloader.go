package josiah

import (
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Downloader struct {
	settings Settings
	Tracker  Tracker
	Model    BibModel
}

type Tracker struct {
	Batches []Batch
}

type Batch struct {
	EndBib   string
	StartBib string
	Filename string
}

func NewDownloader(settings Settings) Downloader {
	d := Downloader{
		settings: settings,
		Model:    NewBibModel(settings),
		Tracker:  Tracker{},
	}
	return d
}

func (d *Downloader) AddBatch(start, end, filename string) Batch {
	batch := Batch{EndBib: end, StartBib: start, Filename: filename}
	d.Tracker.Batches = append(d.Tracker.Batches, batch)
	return batch
}

func (d Downloader) DownloadAll(toc bool) error {
	for _, batch := range d.Tracker.Batches {
		err := d.DownloadBatch(batch, toc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d Downloader) DownloadBatch(batch Batch, toc bool) error {
	if fileExist(batch.Filename) {
		return nil
	}
	bibRange := "[" + batch.StartBib + "," + batch.EndBib + "]"
	idRange := idFromBib(bibRange)
	limit := idRangeLimit(idRange)

	var err error
	var content string
	for true {
		content, err = d.Model.api.Marc(idRange, limit, toc)
		if err == nil {
			break
		}

		http404 := strings.Contains(err.Error(), "Status code 404")
		empty := strings.Contains(content, "Record not found")
		if http404 && empty {
			content = ""
			break
		}

		// TODO - handle this error
		// 2019/12/06 08:28:33 HTTP GET: https://catalog.library.brown.edu/iii/sierra-api/v5/bibs/marc?id=[5470000,5472000]&limit=2001
		// 2019/12/06 08:28:34 HTTP ERROR: {"code":135,"specificCode":0,"httpStatus":500,"name":"External Process Failed","description":"External Process Failed : bib2Marc process failed twice. Error1: External Process Failed : bib2Marc process failed : External Process Failed : bib2Marc process failed. No records exported, check parameters Error2: External Process Failed : bib2Marc process failed : External Process Failed : bib2Marc process failed. No records exported, check parameters"}
		// 2019/12/06 08:28:34 &errors.errorString{s:"Status code 500"}

		retry := strings.Contains(content, "Rate exceeded for endpoint")
		if !retry {
			return err
		}

		log.Printf("Going to sleep for 16 minutes...")
		time.Sleep(16 * time.Minute)
	}

	err = d.writeToFile(batch.Filename, content)
	return err
}

// source https://golangcode.com/writing-to-file/
func (d Downloader) writeToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
