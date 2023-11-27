package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/gonutz/w32/v2"
	"image/color"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

var ApiKey string = "AIzaSyBorAKwVMjlczlsw5HG_GXPAnvv07cJoFc"

type Response struct {
	Location struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
	Accuracy float64 `json:"accuracy"`
}

func updateTime(clock *canvas.Text) {
	formatted := time.Now().Format("PM:03:04")
	clock.Text = formatted
	clock.Refresh()
}

// type blueTheme struct {
// }
//
//	func (b blueTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
//		return theme.DefaultTheme().Icon(name)
//	}
//
//	func (b blueTheme) Font(style fyne.TextStyle) fyne.Resource {
//		return theme.DefaultTheme().Font(style)
//	}
//
//	func (b blueTheme) Size(name fyne.ThemeSizeName) float32 {
//		return theme.DefaultTheme().Size(name)
//	}
//
//	func (b blueTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
//		if name == theme.ColorNameBackground {
//			if variant == theme.VariantLight {
//				return color.NRGBA{R: 81, G: 223, B: 255, A: 255}
//			} else if variant == theme.VariantDark {
//				return color.NRGBA{R: 88, G: 88, B: 86, A: 255}
//			} else {
//				return color.NRGBA{R: 80, G: 80, B: 80, A: 255}
//			}
//		}
//		return theme.DefaultTheme().Color(name, variant)
//	}
func main() {
	err := os.Setenv("FYNE_FONT", "data\\NanumGothic.ttf")
	if err != nil {
		return
	}

	a := app.New()
	w := a.NewWindow("날씨")
	w.Resize(fyne.NewSize(200, 300))
	//a.Settings().SetTheme(&blueTheme{})
	hideConsole()
	url := "https://www.googleapis.com/geolocation/v1/geolocate?key=" + ApiKey
	Ico, err := fyne.LoadResourceFromPath("data\\Icon.png")
	w.SetIcon(Ico)

	Map_res, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Println("Error is", err)
		return
	}
	defer Map_res.Body.Close()

	Map_body, err := io.ReadAll(Map_res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var response Response
	err = json.Unmarshal(Map_body, &response)
	if err != nil {
		fmt.Println(err)
		return
	}

	lat := response.Location.Lat
	lng := response.Location.Lng

	res, err := http.Get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=39a8207fa260e1c283aa32e49aae2177", lat, lng))
	if err != nil {
		fmt.Println("Error is ", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error is", err)
	}

	weather, err := UnmarshalWelcome(body)
	if err != nil {
		fmt.Println("Error is", err)
	}

	code := weather.Weather[0].ID

	Country := canvas.NewText(fmt.Sprintf("%s", weather.Name), color.White)
	Country.TextStyle = fyne.TextStyle{Bold: true}
	Country.TextSize = 50
	Country_id(weather, Country)
	center := container.NewCenter(Country)

	weather_pic := canvas.NewImageFromFile("img\\sun.png")
	weather_pic.FillMode = canvas.ImageFillOriginal

	Back := canvas.NewImageFromFile("img\\BackGround_Sky.png")

	Wind := canvas.NewText(fmt.Sprintf("바람 %.2fms", weather.Wind.Speed), color.White)

	Status := canvas.NewText("", color.White)
	Status.TextSize = 30

	Temputure := canvas.NewText(fmt.Sprintf("%.1f℃", weather.Main.Temp-273.15), color.White)
	Temputure.TextSize = 20

	Min_Max_Temp := canvas.NewText(fmt.Sprintf("최저 %.1f℃ / 최고 %.1f", weather.Main.TempMin-273.15, weather.Main.TempMax-273.15), color.White)
	Min_Max_Temp.TextSize = 15

	MM_Container := container.NewCenter(Min_Max_Temp)

	clock := canvas.NewText("", color.White)
	updateTime(clock)
	go func() {
		for range time.Tick(time.Second) {
			updateTime(clock)
		}
	}()

	Refresh_Ico, err := fyne.LoadResourceFromPath("data\\Reload.png")
	if err != nil {
		fmt.Println(err)
	}
	Refresh := widget.NewToolbar(
		widget.NewToolbarAction(Refresh_Ico, func() {
			code = weather.Weather[0].ID
			NewWeather := Refresh(lat, lng)
			NewCountry := Refresh_Text(NewWeather, Country, Temputure, Min_Max_Temp, Wind)
			Change(weather_pic, code, Status, Back)
			Country_id(NewWeather, NewCountry)
		}))

	Change(weather_pic, code, Status, Back)

	StatusCon := container.NewHBox(Status, Temputure)
	container1 := container.NewVBox(Refresh, center, container.NewCenter(clock), weather_pic, container.NewCenter(StatusCon), container.NewCenter(Wind), MM_Container)
	Main_Con := container.New(layout.NewMaxLayout(), Back, container1)
	w.SetContent(Main_Con)
	w.ShowAndRun()

}

func Refresh(lat, lng float64) Welcome {
	res, err := http.Get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=39a8207fa260e1c283aa32e49aae2177", lat, lng))
	if err != nil {
		fmt.Println(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	weather, err := UnmarshalWelcome(body)
	if err != nil {
		fmt.Println(err)
	}
	return weather
}
func Refresh_Text(weather Welcome, Country *canvas.Text, Temp *canvas.Text, MinMax *canvas.Text, wind *canvas.Text) *canvas.Text {
	Country.Text = fmt.Sprintf("%s", weather.Name)
	Country.Refresh()
	Temp.Text = fmt.Sprintf("%.1f℃", weather.Main.Temp-273.15)
	Temp.Refresh()
	MinMax.Text = fmt.Sprintf("최저 %.1f℃ / 최고 %.1f", weather.Main.TempMin-273.15, weather.Main.TempMax-273.15)
	MinMax.Refresh()
	wind.Text = fmt.Sprintf("바람 %.2fms", weather.Wind.Speed)
	wind.Refresh()

	return Country
}
func IDToKorean(id int64) string {
	idMap := map[int64]string{
		1835848: "서울",
		1838716: "부천시",
		1838524: "부산",
		1838519: "부산",
		1835329: "대구",
		1835327: "대구",
		1840898: "파주",
	}

	korean, ok := idMap[id]
	if !ok {
		korean = "알 수 없음"
	}
	return korean
}

func Country_id(weather Welcome, Country *canvas.Text) {
	korean := IDToKorean(weather.ID)
	Country.Text = korean
	Country.Refresh()
}

func Change(img *canvas.Image, code int64, status *canvas.Text, Back *canvas.Image) {
	format := time.Now().Format("15")
	current, _ := strconv.Atoi(format)
	if code >= 200 && code < 300 {
		img.File = "img\\lightning.png"
		img.Refresh()
		status.Text = "천둥 번개"
		//variant = theme.VariantDark
		Back.File = "img\\BackGround_Dark.png"
		Back.Refresh()

	} else if code >= 300 && code < 700 {
		img.File = "img\\Rain.png"
		status.Text = "비"
		img.Refresh()
		//variant = theme.VariantDark
		Back.File = "img\\BackGround_Dark.png"
		Back.Refresh()

	} else if code >= 700 && code < 800 {
		img.File = "img\\d_cloud.png"
		status.Text = "흐림"
		img.Refresh()
		status.Refresh()
		//variant = theme.VariantDark
		Back.File = "img\\BackGround_Dark.png"
		Back.Refresh()

	} else if code == 800 {
		img.File = "img\\sun.png"
		status.Text = "맑음"
		img.Refresh()
		status.Refresh()
		//variant = theme.VariantDark
		if current >= 18 || current <= 6 {
			img.File = "img\\moon.png"
			img.Refresh()
			//variant = theme.VariantDark
			Back.File = "img\\BackGround_Dark.png"
			Back.Refresh()
		}

	} else if code == 801 || code == 802 {
		fmt.Println("Change cloud")
		img.File = "img\\cloud.png"
		status.Text = "구름 조금"
		img.Refresh()
		status.Refresh()
		//variant = theme.VariantLight
		Back.File = "img\\BackGround_Sky.png"
		Back.Refresh()

		if current >= 18 || current <= 6 {
			img.File = "img\\night_moon.png"
			img.Refresh()
			//variant = theme.VariantDark
			Back.File = "img\\BackGround_Dark.png"
			Back.Refresh()
		}

	} else if code == 803 || code == 804 {
		img.File = "img\\d_cloud.png"
		status.Text = "구름 많음"
		img.Refresh()
		status.Refresh()
		//variant = theme.VariantDark
		Back.File = "img\\BackGround_Dark.png"
		Back.Refresh()
	}
}

func hideConsole() {
	console := w32.GetConsoleWindow()
	if console == 0 {
		return // no console attached
	}
	// If this application is the process that created the console window, then
	// this program was not compiled with the -H=windowsgui flag and on start-up
	// it created a console along with the main application window. In this case
	// hide the console window.
	// See
	// http://stackoverflow.com/questions/9009333/how-to-check-if-the-program-is-run-from-a-console
	_, consoleProcID := w32.GetWindowThreadProcessId(console)
	if w32.GetCurrentProcessId() == consoleProcID {
		w32.ShowWindowAsync(console, w32.SW_HIDE)
	}
}

func UnmarshalWelcome(data []byte) (Welcome, error) {
	var r Welcome
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Weather) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Welcome struct {
	Coord      Coord     `json:"coord"`
	Weather    []Weather `json:"weather"`
	Base       string    `json:"base"`
	Main       Main      `json:"main"`
	Visibility int64     `json:"visibility"`
	Wind       Wind      `json:"wind"`
	Clouds     Clouds    `json:"clouds"`
	Dt         int64     `json:"dt"`
	Sys        Sys       `json:"sys"`
	Timezone   int64     `json:"timezone"`
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Cod        int64     `json:"cod"`
}

type Clouds struct {
	All int64 `json:"all"`
}

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int64   `json:"pressure"`
	Humidity  int64   `json:"humidity"`
}

type Sys struct {
	Type    int64  `json:"type"`
	ID      int64  `json:"id"`
	Country string `json:"country"`
	Sunrise int64  `json:"sunrise"`
	Sunset  int64  `json:"sunset"`
}

type Weather struct {
	ID          int64  `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int64   `json:"deg"`
}
