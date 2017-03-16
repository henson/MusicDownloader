package main

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

const (
	numDowner      = 10   //下载进程
	chanBufferSize = 1000 //通道缓存
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
func main() {
	var URL string
	fmt.Print("Enter url:")
	fmt.Scanf("%s", &URL)
	log.Println("Working, pls wait ...")

	downChan := make(chan Item, chanBufferSize)
	okChan := make(chan int, chanBufferSize)

	if a, b, err := parseURL(URL); err == nil {
		var result []Item
		switch b {
		case "Album":
			var in Album
			result = in.Action(a)
		case "Song":
			var in Song
			result = in.Action(a)
		case "Artist":
			var in Artist
			result = in.Action(a)
		case "ArtistAlbum":
			var in ArtistAlbum
			result = in.Action(a)
		case "Playlist":
			var in Playlist
			result = in.Action(a)
		case "Program":
			var in Program
			result = in.Action(a)
		case "DJradio":
			var in DJradio
			result = in.Action(a)
		}

		for i := 0; i < len(result); i++ {
			downChan <- result[i]
			okChan <- i
		}
	} else {
		log.Println(err.Error())
	}

	for i := 0; i < numDowner; i++ {
		go func() {
			for {
				download(<-downChan, downChan, okChan)
			}
		}()
	}

WAITING:
	if len(okChan) != 0 || len(downChan) != 0 {
		fmt.Printf("=====================================\nDownloading: %v\t\tQueuing: %v\n=====================================\n", len(downChan), len(okChan))
		time.Sleep(10 * time.Second)
		goto WAITING
	}
	log.Println("All tasks finished.")
}

// DJradio 主播电台类型
type DJradio struct {
	Code     int           `json:"code"`
	Count    int           `json:"count"`
	More     bool          `json:"more"`
	Programs []ProgramInfo `json:"programs"`
}

// Song 歌曲类型
type Song struct {
	Code       int           `json:"code"`
	Equalizers []interface{} `json:"equalizers"`
	Songs      []SongInfo    `json:"songs"`
}

// Program 单条主播类型
type Program struct {
	Code    int         `json:"code"`
	Program ProgramInfo `json:"program"`
}

// Artist 歌手类型
type Artist struct {
	Code     int        `json:"code"`
	Artist   ArtistInfo `json:"artist"`
	HotSongs []SongInfo `json:"hotSongs"`
	More     bool       `json:"more"`
}

// ArtistAlbum 歌手专辑类型
type ArtistAlbum struct {
	Code      int         `json:"code"`
	Artist    ArtistInfo  `json:"artist"`
	HotAlbums []AlbumInfo `json:"hotAlbums"`
	More      bool        `json:"more"`
}

// Album 专辑类型
type Album struct {
	Code  int       `json:"code"`
	Album AlbumInfo `json:"album"`
}

// Playlist 歌单类型
type Playlist struct {
	Code   int        `json:"code"`
	Result Collection `json:"result"`
}

//============================

// Collection 类型
type Collection struct {
	AdType                int          `json:"adType"`
	Artists               []ArtistInfo `json:"artists"`
	CloudTrackCount       int          `json:"cloudTrackCount"`
	CommentCount          int          `json:"commentCount"`
	CommentThreadID       string       `json:"commentThreadId"`
	CoverImgID            int          `json:"coverImgId"`
	CoverImgIDStr         string       `json:"coverImgId_str"`
	CoverImgURL           string       `json:"coverImgUrl"`
	CreateTime            int64        `json:"createTime"`
	Creator               UserInfo     `json:"creator"`
	Description           string       `json:"description"`
	HighQuality           bool         `json:"highQuality"`
	ID                    int          `json:"id"`
	Name                  string       `json:"name"`
	NewImported           bool         `json:"newImported"`
	Ordered               bool         `json:"ordered"`
	PlayCount             int          `json:"playCount"`
	Privacy               int          `json:"privacy"`
	ShareCount            int          `json:"shareCount"`
	SpecialType           int          `json:"specialType"`
	Status                int          `json:"status"`
	Subscribed            bool         `json:"subscribed"`
	SubscribedCount       int          `json:"subscribedCount"`
	Subscribers           []string     `json:"subscribers"`
	Tags                  []string     `json:"tags"`
	TotalDuration         int          `json:"totalDuration"`
	TrackCount            int          `json:"trackCount"`
	TrackNumberUpdateTime int64        `json:"trackNumberUpdateTime"`
	TrackUpdateTime       int64        `json:"trackUpdateTime"`
	Tracks                []SongInfo   `json:"tracks"`
	UpdateTime            int64        `json:"updateTime"`
	UserID                int          `json:"userId"`
}

// ProgramInfo 类型
type ProgramInfo struct {
	BdAuditStatus   int        `json:"bdAuditStatus"`
	BlurCoverURL    string     `json:"blurCoverUrl"`
	Buyed           bool       `json:"buyed"`
	Channels        []string   `json:"channels"`
	CommentCount    int        `json:"commentCount"`
	CommentThreadID string     `json:"commentThreadId"`
	CoverURL        string     `json:"coverUrl"`
	CreateTime      int64      `json:"createTime"`
	Description     string     `json:"description"`
	DJ              UserInfo   `json:"dj"`
	Duration        int        `json:"duration"`
	H5Links         string     `json:"h5Links"`
	ID              int        `json:"id"`
	IsPublish       bool       `json:"isPublish"`
	LikedCount      int        `json:"likedCount"`
	ListenerCount   int        `json:"listenerCount"`
	MainSong        SongInfo   `json:"mainSong"`
	MainTrackID     int        `json:"mainTrackId"`
	Name            string     `json:"name"`
	ProgramDesc     string     `json:"programDesc"`
	ProgramFeeType  int        `json:"programFeeType"`
	PubStatus       int        `json:"pubStatus"`
	Radio           RadioInfo  `json:"radio"`
	Reward          bool       `json:"reward"`
	SerialNum       int        `json:"serialNum"`
	ShareCount      int        `json:"shareCount"`
	Songs           []SongInfo `json:"songs"`
	Subscribed      bool       `json:"subscribed"`
	SubscribedCount int        `json:"subscribedCount"`
	TitbitImages    string     `json:"titbitImages"`
	Titbits         string     `json:"titbits"`
	TrackCount      int        `json:"trackCount"`
}

// UserInfo 类型
type UserInfo struct {
	AccountStatus      int      `json:"accountStatus"`
	AuthStatus         int      `json:"authStatus"`
	Authority          int      `json:"authority"`
	AvatarImgID        int      `json:"avatarImgId"`
	AvatarImgIDStr     string   `json:"avatarImgIdStr"`
	AvatarURL          string   `json:"avatarUrl"`
	BackgroundImgID    int      `json:"backgroundImgId"`
	BackgroundImgIDStr string   `json:"backgroundImgIdStr"`
	BackgroundURL      string   `json:"backgroundUrl"`
	Birthday           int64    `json:"birthday"`
	Brand              string   `json:"brand"`
	City               int      `json:"city"`
	DefaultAvatar      bool     `json:"defaultAvatar"`
	Description        string   `json:"description"`
	DetailDescription  string   `json:"detailDescription"`
	DJStatus           int      `json:"djStatus"`
	ExpertTags         []string `json:"expertTags"`
	Followed           bool     `json:"followed"`
	Gender             int      `json:"gender"`
	Mutual             bool     `json:"mutual"`
	Nickname           string   `json:"nickname"`
	Province           int      `json:"province"`
	RemarkName         string   `json:"remarkName"`
	Signature          string   `json:"signature"`
	UserID             int      `json:"userId"`
	UserType           int      `json:"userType"`
	VipType            int      `json:"vipType"`
}

// RadioInfo 类型
type RadioInfo struct {
	Buyed                 bool    `json:"buyed"`
	Category              string  `json:"category"`
	CategoryID            int     `json:"categoryId"`
	CreateTime            int64   `json:"createTime"`
	Desc                  string  `json:"desc"`
	DJ                    string  `json:"dj"`
	Finished              bool    `json:"finished"`
	ID                    int     `json:"id"`
	LastProgramCreateTime int64   `json:"lastProgramCreateTime"`
	LastProgramID         int     `json:"lastProgramId"`
	LastProgramName       string  `json:"lastProgramName"`
	Name                  string  `json:"name"`
	PicURL                string  `json:"picUrl"`
	Price                 float64 `json:"price"`
	ProgramCount          int     `json:"programCount"`
	PurchaseCount         int     `json:"purchaseCount"`
	RadioFeeType          int     `json:"radioFeeType"`
	Rcmdtext              string  `json:"rcmdtext"`
	SubCount              int     `json:"subCount"`
	UnderShelf            bool    `json:"underShelf"`
	Videos                string  `json:"videos"`
}

// ArtistInfo 类型
type ArtistInfo struct {
	AlbumSize   int      `json:"albumSize"`
	Alias       []string `json:"alias"`
	BriefDesc   string   `json:"briefDesc"`
	ID          int      `json:"id"`
	Img1v1ID    int      `json:"img1v1Id"`
	Img1v1IDStr string   `json:"img1v1Id_str"`
	Img1v1URL   string   `json:"img1v1Url"`
	MusicSize   int      `json:"musicSize"`
	Name        string   `json:"name"`
	PicID       int      `json:"picId"`
	PicURL      string   `json:"picUrl"`
	TopicPerson int      `json:"topicPerson"`
	Trans       string   `json:"trans"`
}

// MusicType 类型
type MusicType struct {
	Bitrate     int     `json:"bitrate"`
	DfsID       int64   `json:"dfsId"`
	DfsIDStr    string  `json:"dfsId_str"`
	Extension   string  `json:"extension"`
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	PlayTime    int     `json:"playTime"`
	Size        int64   `json:"size"`
	Sr          int     `json:"sr"`
	VolumeDelta float64 `json:"volumeDelta"`
}

// SongInfo 类型
type SongInfo struct {
	Album           AlbumInfo    `json:"album"`
	Alias           []string     `json:"alias"`
	Artists         []ArtistInfo `json:"artists"`
	Audition        string       `json:"audition"`
	BMusic          MusicType    `json:"bMusic"`
	CommentThreadID string       `json:"commentThreadId"`
	CopyFrom        string       `json:"copyFrom"`
	CopyrightID     int          `json:"copyrightId"`
	Crbt            string       `json:"crbt"`
	DayPlays        int          `json:"dayPlays"`
	Disc            string       `json:"disc"`
	Duration        int          `json:"duration"`
	Fee             int          `json:"fee"`
	Ftype           int          `json:"ftype"`
	HMusic          MusicType    `json:"hMusic"`
	HearTime        int          `json:"hearTime"`
	ID              int          `json:"id"`
	LMusic          MusicType    `json:"lMusic"`
	MMusic          MusicType    `json:"mMusic"`
	Mp3URL          string       `json:"mp3Url"`
	MVID            int          `json:"mvid"`
	Name            string       `json:"name"`
	No              int          `json:"no"`
	PlayedNum       int          `json:"playedNum"`
	Popularity      float64      `json:"popularity"`
	Position        int          `json:"position"`
	Ringtone        string       `json:"ringtone"`
	RtURL           string       `json:"rtUrl"`
	RtURLs          []string     `json:"rtUrls"`
	Rtype           int          `json:"rtype"`
	RURL            string       `json:"rurl"`
	Score           int          `json:"score"`
	Starred         bool         `json:"starred"`
	StarredNum      int          `json:"starredNum"`
	Status          int          `json:"status"`
}

// AlbumInfo 类型
type AlbumInfo struct {
	Alias           []string     `json:"alias"`
	Artist          ArtistInfo   `json:"artist"`
	Artists         []ArtistInfo `json:"artists"`
	BlurPicURL      string       `json:"blurPicUrl"`
	BriefDesc       string       `json:"briefDesc"`
	CommentThreadID string       `json:"commentThreadId"`
	Company         string       `json:"company"`
	CompanyID       int          `json:"companyId"`
	CopyrightID     int          `json:"copyrightId"`
	Description     string       `json:"description"`
	ID              int          `json:"id"`
	Name            string       `json:"name"`
	OnSale          bool         `json:"onSale"`
	Paid            bool         `json:"paid"`
	Pic             int          `json:"pic"`
	PicID           int          `json:"picId"`
	PicIDStr        string       `json:"picId_str"`
	PicURL          string       `json:"picUrl"`
	PublishTime     int64        `json:"publishTime"`
	Size            int          `json:"size"`
	Songs           []SongInfo   `json:"songs"`
	Status          int          `json:"status"`
	SubType         string       `json:"subType"`
	Tags            string       `json:"tags"`
	Atype           string       `json:"type"`
}

// Item 下载文件类型
type Item struct {
	Dir, FileName, FileURL string
}

// Action for DJradio
func (in *DJradio) Action(URL string) (items []Item) {
	gorequest.New().Get(URL).
		Set("cookie", "appver=1.5.2").
		Set("referer", "http://music.163.com/").
		EndStruct(&in)
	var item Item
	for i := 0; i < in.Count; i++ {
		all := in.Count
		item.Dir = in.Programs[i].DJ.Brand + "/"
		if temp := in.Programs[i].MainSong.HMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.Programs[i].MainSong.Name + "(" + strconv.Itoa(in.Programs[i].MainSong.HMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Programs[i].MainSong.HMusic.DfsID, 10))
		} else if temp := in.Programs[i].MainSong.MMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.Programs[i].MainSong.Name + "(" + strconv.Itoa(in.Programs[i].MainSong.MMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Programs[i].MainSong.MMusic.DfsID, 10))
		} else {
			item.FileName = iton(i, all) + "." + in.Programs[i].MainSong.Name + "(" + strconv.Itoa(in.Programs[i].MainSong.LMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Programs[i].MainSong.LMusic.DfsID, 10))
		}
		items = append(items, item)
	}
	return
}

// Action for Song
func (in *Song) Action(URL string) (items []Item) {
	gorequest.New().Get(URL).
		Set("cookie", "appver=1.5.2").
		Set("referer", "http://music.163.com/").
		EndStruct(&in)
	var item Item
	for i := 0; i < len(in.Songs); i++ {
		all := len(in.Songs)
		item.Dir = in.Songs[i].Album.Name + "/"
		if temp := in.Songs[i].HMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.Songs[i].Name + "(" + strconv.Itoa(in.Songs[i].HMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Songs[i].HMusic.DfsID, 10))
		} else if temp := in.Songs[i].MMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.Songs[i].Name + "(" + strconv.Itoa(in.Songs[i].MMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Songs[i].MMusic.DfsID, 10))
		} else {
			item.FileName = iton(i, all) + "." + in.Songs[i].Name + "(" + strconv.Itoa(in.Songs[i].LMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Songs[i].LMusic.DfsID, 10))
		}
		items = append(items, item)
	}
	return
}

// Action for Program
func (in *Program) Action(URL string) (items []Item) {
	gorequest.New().Get(URL).
		Set("cookie", "appver=1.5.2").
		Set("referer", "http://music.163.com/").
		EndStruct(&in)
	var item Item
	item.Dir = in.Program.Name + "/"
	if temp := in.Program.MainSong.HMusic.DfsID; temp != 0 {
		item.FileName = in.Program.MainSong.Name + "(" + strconv.Itoa(in.Program.MainSong.HMusic.Bitrate/1000) + "k).mp3"
		item.FileURL = encryptedID(strconv.FormatInt(in.Program.MainSong.HMusic.DfsID, 10))
	} else if temp := in.Program.MainSong.MMusic.DfsID; temp != 0 {
		item.FileName = in.Program.MainSong.Name + "(" + strconv.Itoa(in.Program.MainSong.MMusic.Bitrate/1000) + "k).mp3"
		item.FileURL = encryptedID(strconv.FormatInt(in.Program.MainSong.MMusic.DfsID, 10))
	} else {
		item.FileName = in.Program.MainSong.Name + "(" + strconv.Itoa(in.Program.MainSong.LMusic.Bitrate/1000) + "k).mp3"
		item.FileURL = encryptedID(strconv.FormatInt(in.Program.MainSong.LMusic.DfsID, 10))
	}
	items = append(items, item)
	return
}

// Action for Artist
func (in *Artist) Action(URL string) (items []Item) {
	gorequest.New().Get(URL).
		Set("cookie", "appver=1.5.2").
		Set("referer", "http://music.163.com/").
		EndStruct(&in)
	var item Item
	for i := 0; i < len(in.HotSongs); i++ {
		all := len(in.HotSongs)
		item.Dir = in.Artist.Name + "/"
		if temp := in.HotSongs[i].HMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.HotSongs[i].Name + "(" + strconv.Itoa(in.HotSongs[i].HMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.HotSongs[i].HMusic.DfsID, 10))
		} else if temp := in.HotSongs[i].MMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.HotSongs[i].Name + "(" + strconv.Itoa(in.HotSongs[i].MMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.HotSongs[i].MMusic.DfsID, 10))
		} else {
			item.FileName = iton(i, all) + "." + in.HotSongs[i].Name + "(" + strconv.Itoa(in.HotSongs[i].LMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.HotSongs[i].LMusic.DfsID, 10))
		}
		items = append(items, item)
	}
	return
}

// Action for ArtistAlbum
func (in *ArtistAlbum) Action(URL string) (items []Item) {
	gorequest.New().Get(URL).
		Set("cookie", "appver=1.5.2").
		Set("referer", "http://music.163.com/").
		EndStruct(&in)
	for i := 0; i < len(in.HotAlbums); i++ {
		tempURL := "http://music.163.com/#/album?id=" + strconv.Itoa(in.HotAlbums[i].ID)
		a, _, _ := parseURL(tempURL)
		var in Album
		tempList := in.Action(a)
		for i := 0; i < len(tempList); i++ {
			items = append(items, tempList[i])
		}
	}
	return
}

// Action for Album
func (in *Album) Action(URL string) (items []Item) {
	gorequest.New().Get(URL).
		Set("cookie", "appver=1.5.2").
		Set("referer", "http://music.163.com/").
		EndStruct(&in)
	var item Item
	for i := 0; i < in.Album.Size; i++ {
		all := in.Album.Size
		item.Dir = in.Album.Name + "/"
		if temp := in.Album.Songs[i].HMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.Album.Songs[i].Name + "(" + strconv.Itoa(in.Album.Songs[i].HMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Album.Songs[i].HMusic.DfsID, 10))
		} else if temp := in.Album.Songs[i].MMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.Album.Songs[i].Name + "(" + strconv.Itoa(in.Album.Songs[i].MMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Album.Songs[i].MMusic.DfsID, 10))
		} else {
			item.FileName = iton(i, all) + "." + in.Album.Songs[i].Name + "(" + strconv.Itoa(in.Album.Songs[i].LMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Album.Songs[i].LMusic.DfsID, 10))
		}
		items = append(items, item)
	}
	return
}

// Action for Playlist
func (in *Playlist) Action(URL string) (items []Item) {
	gorequest.New().Get(URL).
		Set("cookie", "appver=1.5.2").
		Set("referer", "http://music.163.com/").
		EndStruct(&in)
	var item Item
	for i := 0; i < in.Result.TrackCount; i++ {
		all := in.Result.TrackCount
		item.Dir = in.Result.Name + "/"
		if temp := in.Result.Tracks[i].HMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.Result.Tracks[i].Name + "(" + strconv.Itoa(in.Result.Tracks[i].HMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Result.Tracks[i].HMusic.DfsID, 10))
		} else if temp := in.Result.Tracks[i].MMusic.DfsID; temp != 0 {
			item.FileName = iton(i, all) + "." + in.Result.Tracks[i].Name + "(" + strconv.Itoa(in.Result.Tracks[i].MMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Result.Tracks[i].MMusic.DfsID, 10))
		} else {
			item.FileName = iton(i, all) + "." + in.Result.Tracks[i].Name + "(" + strconv.Itoa(in.Result.Tracks[i].LMusic.Bitrate/1000) + "k).mp3"
			item.FileURL = encryptedID(strconv.FormatInt(in.Result.Tracks[i].LMusic.DfsID, 10))
		}
		items = append(items, item)
	}
	return
}

func parseURL(URL string) (sURL string, action string, err error) {
	URL = strings.Replace(URL, "/#", "", -1)
	u, err := url.Parse(URL)
	if err != nil {
		return
	}
	if u.Host != "music.163.com" {
		err = errors.New("URL is not from music.163.com")
		return
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return
	}
	var prefix, id string
	prefix = "http://music.163.com/api"
	id = m["id"][0]

	switch u.Path {
	case "/album":
		sURL = prefix + "/album/" + id
		action = "Album"
	case "/song":
		sURL = prefix + "/song/detail/?id=" + id + "&ids=%5B" + id + "%5D"
		action = "Song"
	case "/artist":
		sURL = prefix + "/artist/" + id
		action = "Artist"
	case "/artist/album":
		sURL = prefix + "/artist/albums/" + id + "?offset=0&limit=1000"
		action = "ArtistAlbum"
	case "/playlist":
		sURL = prefix + "/playlist/detail?id=" + id + "&ids=%5B" + id + "%5D"
		action = "Playlist"
	case "/program":
		sURL = prefix + "/dj/program/detail?id=" + id + "&ids=%5B" + id + "%5D"
		action = "Program"
	case "/djradio":
		sURL = prefix + "/dj/program/byradio?radioId=" + id + "&asc=1&limit=1000"
		action = "DJradio"
	}
	return
}

func encryptedID(id string) (URLString string) {
	byte1 := []byte(id)
	byte2 := []byte("3go8&$8*3*3h0k(2)2")
	for i, v := range byte1 {
		byte1[i] = v ^ byte2[i%len(byte2)]
	}
	h := md5.New()
	h.Write(byte1)
	encodeString := base64.StdEncoding.EncodeToString(h.Sum(nil))
	encodeString = strings.Replace(encodeString, "/", "_", -1)
	encodeString = strings.Replace(encodeString, "+", "-", -1)

	URLString = "http://p" + strconv.Itoa(RandInt(1, 3)) + ".music.126.net/" + encodeString + "/" + id + ".mp3"
	return
}

// RandInt 生成min~max范围内随机整数
func RandInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}

//Exist 是否存在
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func download(i Item, retry chan Item, ok chan int) {
	saveFile := i.Dir + clearSymbol(i.FileName)
	if Exist(saveFile) {
		log.Printf("%v is exist.\n", saveFile)
		<-ok
		return
	}
	err := os.MkdirAll(i.Dir, 0777)
	if err != nil {
		log.Print("downloadFile[1]:" + i.FileURL + " ====>>> " + err.Error())
		retry <- i
		return
	}
	resp, err := http.Get(i.FileURL)
	if err != nil {
		log.Print("downloadFile[2]:" + i.FileURL + " ====>>> " + err.Error())
		retry <- i
		return
	}
	defer resp.Body.Close()
	fd, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print("downloadFile[3]:" + i.FileURL + " ====>>> " + err.Error())
		retry <- i
		return
	}
	err = ioutil.WriteFile(saveFile, fd, 0666)
	if err != nil {
		log.Print("downloadFile[4]:" + i.FileURL + " ====>>> " + err.Error())
		retry <- i
		return
	}
	log.Println(saveFile + " downloaded.")
	<-ok
}

func clearSymbol(in string) (out string) {
	out = strings.Replace(in, ":", "", -1)
	out = strings.Replace(out, ",", "", -1)
	out = strings.Replace(out, "\"", "", -1)
	return
}

// iton parse the int to NO. with zero
func iton(no, all int) string {
	return fmt.Sprintf("%0"+strconv.Itoa(len(strconv.Itoa(all)))+"d", no+1)
}
