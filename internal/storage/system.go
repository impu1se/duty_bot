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

func (l *System) ReadDutyCSV(file string) (map[string][][]string, error) {
	dutyList := make(map[string][][]string)
	csvFile, err := os.Open(file)

	if err != nil {
		l.logger.Error(err.Error())
		return dutyList, err
	}
	defer csvFile.Close()
	l.logger.Info("Successfully Opened CSV file")

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		l.logger.Error(err.Error())
		return dutyList, err
	}

	for _, nameCommand := range csvLines[0][1:] {
		dutyList[nameCommand] = make([][]string, len(csvLines), len(csvLines))
	}
	for j, line := range csvLines[1:] {
		for i, nameCommand := range csvLines[0][1:] {
			dutyList[nameCommand][j] = append(dutyList[nameCommand][j], line[0], line[i+1])
		}
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
