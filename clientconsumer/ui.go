package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"time"
)

const (
	TempBtnText1  = "  Subscribe Temperature  "
	TempBtnText2  = "Unsubscribe Temperature"
	HumidBtnText1 = "  Subscribe Humid  "
	HumidBtnText2 = "Unsubscribe Humid"

	Tem = 0
	Hum = 1
)

var (
	tempButton  *widget.Button
	humidButton *widget.Button

	tempBtnClicked  = false
	humidBtnClicked = false

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
	w := a.NewWindow("ENS_GO_Consumer ***Client***")
	w.SetOnClosed(beforeClose)

	InitNetCfg()
	if err, addr := InitConn(); err != nil {
		return err
	} else {
		_ = logs.Prepend("Connected to Server: " + addr)
	}
	restore()

	// Button
	tempButton = widget.NewButton(TempBtnText1, func() {
		btnClicked(Tem)
	})

	humidButton = widget.NewButton(HumidBtnText1, func() {
		btnClicked(Hum)
	})
	buttons := container.NewHBox(
		tempButton,
		widget.NewSeparator(),
		humidButton,
	)

	// Display Bar for Temperature and Humid
	bindTemp = binding.NewFloat()
	bindHumid = binding.NewFloat()
	tempLabel := widget.NewLabel("Temperature")
	tempLabel.TextStyle.Bold = true
	tempBar := widget.NewProgressBarWithData(bindTemp)
	tempBar.TextFormatter = func() string {
		return fmt.Sprintf("%d ℃", int(tempBar.Value*100))
	}
	humidLabel := widget.NewLabel("Humidity")
	humidLabel.TextStyle.Bold = true
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
		buttons,
		widget.NewSeparator(),

		tempLabel, tempBar,
		humidLabel, humidBar,

		widget.NewSeparator(),
	)

	w.SetContent(container.NewGridWithRows(
		2,
		up,
		logPanel,
	))

	w.Resize(fyne.Size{Width: 400, Height: 450})
	w.SetFixedSize(true)

	go WaitForUpdate(update)

	w.ShowAndRun()

	return nil
}

func restore() {
	UnSubscribeEvent(EventTemperature)
	UnSubscribeEvent(EventHumid)
}

func update(temp_ int, humid_ int, log string) {
	if temp_ != 0 {
		if tempBtnClicked == false {
			tempBtnClicked = true
			tempButton.SetText(TempBtnText2)
		}
		if err := bindTemp.Set(float64(temp_) / 100.0); err != nil {
			fmt.Println("set temperature error")
		}
	}
	if humid_ != 0 {
		if humidBtnClicked == false {
			humidBtnClicked = true
			humidButton.SetText(HumidBtnText2)
		}
		if err := bindHumid.Set(float64(humid_) / 100.0); err != nil {
			fmt.Println("set temperature error")
		}
	}
	if err := logs.Prepend(log); err != nil {
		fmt.Println("log error")
	}
}

func btnClicked(t int) {
	switch t {
	case Tem:
		if tempBtnClicked {
			tempBtnClicked = !tempBtnClicked
			tempButton.SetText(TempBtnText1)
			UnSubscribeEvent(EventTemperature)
			log("Unsubscribe Event: " + EventTemperature)
		} else {
			tempBtnClicked = !tempBtnClicked
			tempButton.SetText(TempBtnText2)
			SubscribeEvent(EventTemperature)
			log("Subscribe Event: " + EventTemperature)
		}
	case Hum:
		if humidBtnClicked {
			humidBtnClicked = !humidBtnClicked
			humidButton.SetText(HumidBtnText1)
			UnSubscribeEvent(EventHumid)
			log("Unsubscribe Event: " + EventHumid)
		} else {
			humidBtnClicked = !humidBtnClicked
			humidButton.SetText(HumidBtnText2)
			SubscribeEvent(EventHumid)
			log("Subscribe Event: " + EventHumid)
		}
	}
}

func log(s string) {
	now := time.Now()
	_ = logs.Prepend(now.Format("15:04:05") + " " + s)
}

func beforeClose() {
	ensConn.Close()
}
