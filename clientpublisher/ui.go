package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"time"
)

var (
	bindTemp  binding.Float
	bindHumid binding.Float

	logs = binding.BindStringList(
		&[]string{},
	)
)

func Run() error {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	a.SetIcon(theme.FyneLogo())
	w := a.NewWindow("ENS_GO_Consumer ***Publisher***")
	w.SetOnClosed(beforeClose)

	InitNetCfg()
	if err, addr := InitConn(); err != nil {
		return err
	} else {
		_ = logs.Prepend("Connected to Server: " + addr)
	}

	// Display Bar for Temperature and Humid

	bindTemp = binding.NewFloat()
	bindHumid = binding.NewFloat()
	tempLabel := widget.NewLabel("Temperature")
	tempLabel.TextStyle.Bold = true
	tempSlide := widget.NewSliderWithData(0, 1, bindTemp)
	tempSlide.Step = 0.01
	tempSlide.OnChanged = publishTemp
	tempBar := widget.NewProgressBarWithData(bindTemp)
	tempBar.TextFormatter = func() string {
		return fmt.Sprintf("%d ℃", int(tempBar.Value*100))
	}

	humidLabel := widget.NewLabel("Humidity")
	humidLabel.TextStyle.Bold = true
	humidSlide := widget.NewSliderWithData(0, 1, bindHumid)
	humidSlide.Step = 0.01
	humidSlide.OnChanged = publishHumid
	humidBar := widget.NewProgressBarWithData(bindHumid)

	// Logs
	logger := widget.NewListWithData(logs,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})
	title := widget.NewLabel("Event Logs")
	title.TextStyle.Bold = true

	logPanel := container.NewBorder(title, nil, nil, nil, logger)

	up := container.NewVBox(
		tempLabel, tempSlide, tempBar,
		humidLabel, humidSlide, humidBar,
		widget.NewSeparator(),
	)

	w.SetContent(container.NewGridWithRows(
		2,
		up,
		logPanel,
	))

	w.Resize(fyne.Size{Width: 400, Height: 450})
	w.SetFixedSize(true)

	w.ShowAndRun()

	return nil
}

func publishTemp(v float64) {
	s := strconv.Itoa(int(v * 100))
	PublishEvent(EventTemperature, s)
	_ = bindTemp.Set(v)
	log("Publish Event: " + EventTemperature + ", Value: " + s)
}

func publishHumid(v float64) {
	s := strconv.Itoa(int(v * 100))
	PublishEvent(EventHumid, s)
	_ = bindHumid.Set(v)
	log("Publish Event: " + EventHumid + ", Value: " + s)
}

func log(s string) {
	now := time.Now()
	_ = logs.Prepend(now.Format("15:04:05") + " " + s)
}

func beforeClose() {
	ensConn.Close()
}
