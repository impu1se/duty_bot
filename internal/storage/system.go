package storage

import (
	"encoding/csv"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

const dataDir = "duty_list/"

type System struct {
	logger *zap.Logger
}

func NewLoader(logger *zap.Logger) *System {
	err := os.Mkdir(dataDir, os.FileMode(0777))
	if err != nil {
		logger.Error("can't create data dir")
	}

	return &System{logger: logger}
}

func (l *System) ReadDutyCSV(file string) ([][]string, error) {
	csvFile, err := os.Open(file)

	if err != nil {
		l.logger.Error(err.Error())
		return [][]string{}, err
	}
	defer csvFile.Close()
	l.logger.Info("Successfully Opened CSV file")

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		l.logger.Error(err.Error())
		return [][]string{}, err
	}

	dutyList := make([][]string, len(csvLines), len(csvLines))
	for i, line := range csvLines[1:] {
		dutyList[i] = append(dutyList[i], line[0], line[1])
	}

	return dutyList, nil
}

func (l *System) ClearDir(pattern string) error {
	files, err := filepath.Glob(dataDir + pattern)
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}
