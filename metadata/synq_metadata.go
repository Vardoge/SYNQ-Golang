package metadata

type MetaData struct {
	Version          string       `json:"metadata_version"`
	Title            LanguageList `json:"title"`
	Description      LanguageList `json:"description"`
	Year             int          `json:"production_year"`
	ReleaseYear      int          `json:"release_year"`
	Type             string       `json:"type"`
	Series           Series       `json:"series,omitempty"`
	Genres           []string     `json:"genres"`
	Credits          []Credit     `json:"credits"`
	Regional         bool         `json:"regional_content"`
	Rating           string       `json:"parental_rating"`
	Ratio            string       `json:"aspect_ratio,omitempty"`
	Duration         string       `json:"expected_duration,omitempty"`
	Countries        []string     `json:"country_of_origin"`
	ReleaseDate      string       `json:"first_release_date,omitempty"`
	OriginalLanguage string       `json:"original_language,omitempty"`
	Studio           string       `json:"studio,omitempty"`
	ImdbUrl          string       `json:"imdb_url,omitempty"`
	Ratings          []Rating     `json:"ratings,omitempty"`
	MetadataScore    string       `json:"metadata_score,omitempty"`
	Awards           string       `json:"awards_and_recognitions,omitempty"`
}

type Series struct {
	Episode      int    `json:"episode_number,omitempty"`
	Season       int    `json:"season,omitempty"`
	ExternalId   string `json:"external_id,omitempty"`
	InternalId   string `json:"internal_id,omitempty"`
	EpisodeCount int    `json:"episodes_in_season,omitempty"`
}

type Credit struct {
	Name     string `json:"name"`
	Function string `json:"role"`
}

type Language map[string]string
type LanguageList map[string]Language

type ImageData struct {
	Type        string `json:"type"`
	Orientation string `json:"orientation"`
	Language    string `json:"language,omitempty"`
	File        string `json:"org_file"`
}

type AkkaXMLAsset struct {
	ContentId string `json:"content_id,omitempty"`
	MetaData
	Images []ImageData   `json:"images,omitempty"`
	Rights []VideoRights `json:"rights,omitempty"`
	MultiformData
}

// NOTE : This is probably not an all inclusive list of rights but reflects what we have samples for
type VideoRights struct {
	ValidFrom string   `json:"valid_from"`
	ValidTo   string   `json:"valid_to"`
	Unlimited bool     `json:"unlimited"`
	Devices   []string `json:"devices"`
}

type Rating struct {
	Country string `json:"country"`
	Content string `json:"content"`
}

type MultiformData struct {
	OtherInformation LanguageList `json:"other_information,omitempty"`
	Tags             []string     `json:"tags,omitempty"`
}
