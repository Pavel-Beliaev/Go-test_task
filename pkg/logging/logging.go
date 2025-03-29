package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

// writerHook - структура для хука,
// который перенаправляет логи в несколько потоков
// (например, в файл и консоль одновременно).
type writerHook struct {
	Writer    []io.Writer    // Список мест, куда будут записываться логи
	LogLevels []logrus.Level // Уровни логирования, которые будут обрабатываться
}

// Fire вызывается Logrus при создании новой записи лога.
// Он записывает лог во все указанные Writer.
func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String() // Преобразуем запись в строку
	if err != nil {
		return err
	}
	// Записываем лог в каждый указанный Writer (файл, консоль и т.д.)
	for _, w := range hook.Writer {
		w.Write([]byte(line))
	}
	return err
}

// Levels возвращает список уровней логирования, на которые реагирует хук.
func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

var e *logrus.Entry // Глобальная переменная для логгера

// Logger - обертка вокруг logrus.Entry, добавляющая удобные методы.
type Logger struct {
	*logrus.Entry
}

// GetLogger возвращает текущий логгер.
func GetLogger() Logger {
	return Logger{e}
}

// GetLoggerWithField создает новый логгер с дополнительным полем.
func (l *Logger) GetLoggerWithField(k string, v interface{}) Logger {
	return Logger{l.WithField(k, v)}
}

// init автоматически вызывается при загрузке пакета и настраивает логгер.
func init() {
	l := logrus.New()       // Создаем новый экземпляр Logrus
	l.SetReportCaller(true) // Включаем отображение файла и строки вызова

	// Настройка форматтера
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			filename := path.Base(f.File) // Оставляем только имя файла
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
		DisableColors: false, // Включаем цветной вывод (если терминал поддерживает)
		FullTimestamp: true,  // Включаем полные временные метки
	}

	// Создаем папку для логов, если она не существует
	err := os.MkdirAll("../logs", 0755)
	if err != nil {
		panic(err)
	}

	// Открываем файл logs/all.log для записи логов
	allFile, err := os.OpenFile("../logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	// Отключаем стандартный вывод логов (они будут идти только в хук)
	l.SetOutput(io.Discard)

	// Добавляем хук, который записывает логи в файл и консоль
	l.AddHook(&writerHook{
		Writer:    []io.Writer{allFile, os.Stdout}, // Логи будут в файле и в консоли
		LogLevels: logrus.AllLevels,                // Логи всех уровней
	})

	// Устанавливаем минимальный уровень логирования (самый подробный - Trace)
	l.SetLevel(logrus.TraceLevel)

	// Создаем глобальный экземпляр логгера
	e = logrus.NewEntry(l)
}
