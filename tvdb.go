package tvdb

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// PipeList type representing pipe-separated string values.
type PipeList []string

// UnmarshalXML unmarshals an XML element with string value into a pip-separated list of strings.
func (pipeList *PipeList) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) (err error) {
	content := ""
	if err = decoder.DecodeElement(&content, &start); err != nil {
		return err
	}

	*pipeList = strings.Split(strings.Trim(content, "|"), "|")
	return
}

// Episode represents a TV show episode on TheTVDB.
type Episode struct {
	ID                    uint64   `xml:"id"`
	CombinedEpisodeNumber string   `xml:"Combined_episodenumber"`
	CombinedSeason        uint64   `xml:"Combined_season"`
	DvdChapter            string   `xml:"DVD_chapter"`
	DvdDiscID             string   `xml:"DVD_discid"`
	DvdEpisodeNumber      string   `xml:"DVD_episodenumber"`
	DvdSeason             string   `xml:"DVD_season"`
	Director              PipeList `xml:"Director"`
	EpImgFlag             string   `xml:"EpImgFlag"`
	EpisodeName           string   `xml:"EpisodeName"`
	EpisodeNumber         uint64   `xml:"EpisodeNumber"`
	FirstAired            string   `xml:"FirstAired"`
	GuestStars            string   `xml:"GuestStars"`
	ImdbID                string   `xml:"IMDB_ID"`
	Language              string   `xml:"Language"`
	Overview              string   `xml:"Overview"`
	ProductionCode        string   `xml:"ProductionCode"`
	Rating                string   `xml:"Rating"`
	RatingCount           string   `xml:"RatingCount"`
	SeasonNumber          uint64   `xml:"SeasonNumber"`
	Writer                PipeList `xml:"Writer"`
	AbsoluteNumber        string   `xml:"absolute_number"`
	Filename              string   `xml:"filename"`
	LastUpdated           string   `xml:"lastupdated"`
	SeasonID              uint64   `xml:"seasonid"`
	SeriesID              uint64   `xml:"seriesid"`
	ThumbAdded            string   `xml:"thumb_added"`
	ThumbHeight           string   `xml:"thumb_height"`
	ThumbWidth            string   `xml:"thumb_width"`
}

// Series represents TV show on TheTVDB.
type Series struct {
	ID            uint64   `xml:"id"`
	Actors        PipeList `xml:"Actors"`
	AirsDayOfWeek string   `xml:"Airs_DayOfWeek"`
	AirsTime      string   `xml:"Airs_Time"`
	ContentRating string   `xml:"ContentRating"`
	FirstAired    string   `xml:"FirstAired"`
	Genre         PipeList `xml:"Genre"`
	ImdbID        string   `xml:"IMDB_ID"`
	Language      string   `xml:"Language"`
	Network       string   `xml:"Network"`
	NetworkID     string   `xml:"NetworkID"`
	Overview      string   `xml:"Overview"`
	Rating        string   `xml:"Rating"`
	RatingCount   string   `xml:"RatingCount"`
	Runtime       string   `xml:"Runtime"`
	SeriesID      string   `xml:"SeriesID"`
	SeriesName    string   `xml:"SeriesName"`
	Status        string   `xml:"Status"`
	Added         string   `xml:"added"`
	AddedBy       string   `xml:"addedBy"`
	Banner        string   `xml:"banner"`
	Fanart        string   `xml:"fanart"`
	LastUpdated   string   `xml:"lastupdated"`
	Poster        string   `xml:"poster"`
	Zap2ItID      string   `xml:"zap2it_id"`
	Seasons       map[uint64][]*Episode
}

// SeriesList represents a list of TV shows, often used for returning search results.
type SeriesList struct {
	Series []*Series `xml:"Series"`
}

// EpisodeList represents a list of TV show episodes.
type EpisodeList struct {
	Episodes []*Episode `xml:"Episode"`
}

type TVDB struct {
	APIKey string
}

func NewTVDB(apiKey string) *TVDB {
	return &TVDB{
		APIKey: apiKey,
	}
}

// GetSeries gets a list of TV series by name, by performing a simple search.
func (t *TVDB) GetSeries(name string) (seriesList SeriesList, err error) {
	url := fmt.Sprintf("http://thetvdb.com/api/GetSeries.php?seriesname=%v", url.QueryEscape(name))
	response, err := http.Get(url)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = xml.Unmarshal(data, &seriesList)
	return
}

// GetSeriesByID gets a TV series by ID.
func (t *TVDB) GetSeriesByID(id uint64) (series *Series, err error) {
	url := fmt.Sprintf("http://thetvdb.com/api/%v/series/%v/en.xml", t.APIKey, id)
	response, err := http.Get(url)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	seriesList := SeriesList{}
	if err = xml.Unmarshal(data, &seriesList); err != nil {
		return
	}

	if len(seriesList.Series) != 1 {
		err = errors.New("incorrect number of series")
		return
	}

	series = seriesList.Series[0]
	return
}

// GetSeriesByIMDBID gets series from IMDb's ID.
func (t *TVDB) GetSeriesByIMDBID(id string) (series *Series, err error) {
	url := fmt.Sprintf("http://thetvdb.com/api/GetSeriesByRemoteID.php?imdbid=%v", id)
	response, err := http.Get(url)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	seriesList := SeriesList{}
	if err = xml.Unmarshal(data, &seriesList); err != nil {
		return
	}

	if len(seriesList.Series) != 1 {
		err = errors.New("incorrect number of series")
		return
	}

	series = seriesList.Series[0]
	return
}

// GetDetail gets more detail for all TV shows in a list.
func (t *TVDB) GetSeriesListDetail(seriesList *SeriesList) (err error) {
	for seriesIndex := range seriesList.Series {
		if err = t.GetSeriesDetail(seriesList.Series[seriesIndex]); err != nil {
			return
		}
	}
	return
}

// GetDetail gets more detail for a TV show, including information on it's episodes.
func (t *TVDB) GetSeriesDetail(series *Series) (err error) {
	url := fmt.Sprintf("http://thetvdb.com/api/%v/series/%v/all/en.xml", t.APIKey, strconv.FormatUint(series.ID, 10))
	response, err := http.Get(url)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if err = xml.Unmarshal(data, series); err != nil {
		return
	}

	episodeList := EpisodeList{}
	if err = xml.Unmarshal(data, &episodeList); err != nil {
		return
	}

	if series.Seasons == nil {
		series.Seasons = make(map[uint64][]*Episode)
	}

	for _, episode := range episodeList.Episodes {
		series.Seasons[episode.SeasonNumber] = append(series.Seasons[episode.SeasonNumber], episode)
	}
	return
}

var reSearchSeries = regexp.MustCompile(`(?P<before><a href="/\?tab=series&amp;id=)(?P<seriesId>\d+)(?P<after>\&amp;lid=\d*">)`)

// SearchSeries searches for TV shows by name, using the more sophisticated
// search on TheTVDB's homepage. This is the recommended search method.
func (t *TVDB) SearchSeries(name string, maxResults int) (seriesList SeriesList, err error) {
	url := fmt.Sprintf("http://thetvdb.com/?string=%v&searchseriesid=&tab=listseries&function=Search",
		url.QueryEscape(name))
	response, err := http.Get(url)
	if err != nil {
		return
	}

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	groups := reSearchSeries.FindAllSubmatch(buf, -1)
	doneSeriesIDs := make(map[uint64]struct{})

	for _, group := range groups {
		seriesID := uint64(0)
		var series *Series
		seriesID, err = strconv.ParseUint(string(group[2]), 10, 64)

		if _, ok := doneSeriesIDs[seriesID]; ok {
			continue
		}

		if err != nil {
			return
		}

		series, err = t.GetSeriesByID(seriesID)
		if err != nil {
			// Some series can't be found, so we will ignore these.
			if _, ok := err.(*xml.SyntaxError); ok {
				err = nil

				continue
			} else {
				return
			}
		}

		seriesList.Series = append(seriesList.Series, series)
		doneSeriesIDs[seriesID] = struct{}{}

		if len(seriesList.Series) == maxResults {
			break
		}
	}
	return
}
