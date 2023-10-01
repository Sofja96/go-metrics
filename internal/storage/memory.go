package storage

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"os"
)

func saveStorageToFile(s MemStorage, filePath string) error {
	var metrics AllMetrics
	metrics.Counter = s.counterData
	metrics.Gauge = s.gaugeData

	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0666)

}

func Storing(s MemStorage, filePath string, storeInterval int) {
	dir, _ := path.Split(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0666)
		if err != nil {
			fmt.Println(err)
		}
	}
	pollTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer pollTicker.Stop()
	for range pollTicker.C {
		saveStorageToFile(s, filePath)
	}
}

func LoadStorageFromFile(s MemStorage, filePath string) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}

	var data AllMetrics
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Println(err)
	}

	if len(data.Counter) != 0 {
		s.UpdateCounterData(data.Counter)
	}
	if len(data.Gauge) != 0 {
		s.UpdateGaugeData(data.Gauge)
	}
}
