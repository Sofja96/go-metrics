package memory

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadStorageFromFile(t *testing.T) {
	filePath := "./test_data.json"

	t.Run("successful load", func(t *testing.T) {
		data := AllMetrics{
			Counter: map[string]Counter{"requests": 100},
			Gauge:   map[string]Gauge{"temperature": 23.5},
		}

		fileContent, err := json.Marshal(data)
		assert.NoError(t, err, "Ошибка при маршаллинге данных: %v", err)

		err = os.WriteFile(filePath, fileContent, 0644)
		assert.NoError(t, err, "Ошибка при создании файла: %v", err)

		defer os.Remove(filePath)

		storage := &MemStorage{
			gaugeData:   make(map[string]Gauge),
			counterData: make(map[string]Counter),
		}

		err = LoadStorageFromFile(storage, filePath)
		assert.NoError(t, err, "Ошибка при загрузке данных из файла: %v", err)

		assert.Equal(t, Counter(100), storage.counterData["requests"])
		assert.Equal(t, Gauge(23.5), storage.gaugeData["temperature"])
	})

	t.Run("file not found", func(t *testing.T) {
		storage := &MemStorage{
			gaugeData:   make(map[string]Gauge),
			counterData: make(map[string]Counter),
		}
		err := LoadStorageFromFile(storage, "./nonexistent_file.json")
		assert.Errorf(t, err, "error read and load data from file: %v", err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{"counter": {"requests": 100}, "gauge": "invalid_value"}`)
		err := os.WriteFile(filePath, invalidJSON, 0644)
		assert.NoError(t, err)

		defer os.Remove(filePath)

		storage := &MemStorage{
			gaugeData:   make(map[string]Gauge),
			counterData: make(map[string]Counter),
		}
		err = LoadStorageFromFile(storage, filePath)
		assert.Errorf(t, err, "error read and load data from file: %v", fmt.Errorf("json: cannot unmarshal string into Go struct field AllMetrics.Gauge of type map[string]memory.Gauge"))
	})

	t.Run("empty file", func(t *testing.T) {
		err := os.WriteFile(filePath, []byte{}, 0644)
		assert.NoError(t, err)

		defer os.Remove(filePath)

		storage := &MemStorage{
			gaugeData:   make(map[string]Gauge),
			counterData: make(map[string]Counter),
		}
		err = LoadStorageFromFile(storage, filePath)
		assert.Errorf(t, err, "Ошибка при загрузке данных из пустого файла: %v", err)

		assert.Empty(t, storage.counterData)
		assert.Empty(t, storage.gaugeData)
	})
}

func TestSaveStorageToFile(t *testing.T) {
	storage := &MemStorage{
		gaugeData:   map[string]Gauge{"temperature": 23.5},
		counterData: map[string]Counter{"requests": 100},
	}

	filePath := "./test_storage.json"
	defer os.Remove(filePath)

	err := saveStorageToFile(storage, filePath)
	assert.NoError(t, err, "Ошибка при записи данных в файл")

	fileContent, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Ошибка при чтении файла")

	var metrics AllMetrics
	err = json.Unmarshal(fileContent, &metrics)
	assert.NoError(t, err, "Ошибка при разборе JSON из файла")

	assert.Equal(t, storage.counterData, metrics.Counter, "Метрики counter не совпадают")
	assert.Equal(t, storage.gaugeData, metrics.Gauge, "Метрики gauge не совпадают")

	var expectedMetrics AllMetrics
	expectedMetrics.Counter = storage.counterData
	expectedMetrics.Gauge = storage.gaugeData
	expectedData, _ := json.MarshalIndent(expectedMetrics, "", "   ")

	assert.JSONEq(t, string(expectedData), string(fileContent), "Содержимое файла не соответствует ожидаемому JSON с отступами")
}

func TestSaveStorageToFile_ErrorOnWrite(t *testing.T) {
	storage := &MemStorage{
		gaugeData:   map[string]Gauge{"temperature": 23.5},
		counterData: map[string]Counter{"requests": 100},
	}
	filePath := "/root/test_storage.json"
	err := saveStorageToFile(storage, filePath)

	assert.Errorf(t, err, "Ожидалась ошибка при записи в файл, но её не произошло")
}
