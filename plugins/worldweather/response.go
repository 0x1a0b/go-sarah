package worldweather

// `{ "data": { "error": [ {"msg": "Unable to find any matching weather location to the query submitted!" } ] }}`
type ErrorDescription struct {
	Message string `json:"msg"`
}

type CommonData struct {
	Error []*ErrorDescription `json:"error"`
}

func (data *CommonData) HasError() bool {
	return len(data.Error) > 0
}

// https://developer.worldweatheronline.com/api/docs/local-city-town-weather-api.aspx#data_element
type LocalWeatherResponse struct {
	Data *WeatherData `json:"data"`
}

type WeatherData struct {
	CommonData
	Request          []*Request          `json:"request"`
	CurrentCondition []*CurrentCondition `json:"current_condition"`
	Weather          []*Weather          `json:"weather"`
}

type Request struct {
	Type  string
	Query string
}

type CurrentCondition struct {
	ObservationTime    string                `json:"observation_time"`
	Temperature        string                `json:"temp_C"`
	FeelingTemperature string                `json:"FeelsLikeC"`
	WindSpeed          string                `json:"windspeedKmph"`
	WindDirection      string                `json:"winddirDegree"`
	WeatherCode        string                `json:"weatherCode"`
	WeatherIcon        []*WeatherIcon        `json:"weatherIconUrl"`
	Description        []*WeatherDescription `json:"weatherDesc"`
	Precipitation      string                `json:"precpMM"`
	Humidity           string                `json:"humidity"`
	Visibility         string                `json:"visibility"`
	Pressure           string                `json:"pressure"`
	CloudCover         string                `json:"cloudcocver"`
}

type WeatherIcon struct {
	URL string `json:"value"`
}

type WeatherDescription struct {
	Content string `json:"value"`
}

type Weather struct {
	Astronomy []*Astronomy     `json:"astronomy"`
	Date      string           `json:"date"` // 2016-09-04
	MaxTempC  string           `json:"maxTempC"`
	MaxTempF  string           `json:"maxTempF"`
	MinTempC  string           `json:"minTempC"`
	MinTempF  string           `json:"minTempF"`
	UV        string           `json:"uvindex"`
	Hourly    []*HourlyWeather `json:"hourly"`
}

type HourlyWeather struct {
	Time               string                `json:"time"`
	Temperature        string                `json:"tempC"` // not temp_C
	FeelingTemperature string                `json:"FeelsLikeC"`
	WindSpeed          string                `json:"windspeedKmph"`
	WindDirection      string                `json:"winddirDegree"`
	WeatherCode        string                `json:"weatherCode"`
	WeatherIcon        []*WeatherIcon        `json:"weatherIconUrl"`
	Description        []*WeatherDescription `json:"weatherDesc"`
	Precipitation      string                `json:"precpMM"`
	Humidity           string                `json:"humidity"`
	Visibility         string                `json:"visibility"`
	Pressure           string                `json:"pressure"`
	CloudCover         string                `json:"cloudcocver"`
}

// TODO
type Astronomy struct {
	Sunrise  string `json:"sunrise"`
	Sunset   string `json:"sunset"`
	MoonRise string `json:"moonrise"`
	MoonSet  string `json:"moonset"`
}
