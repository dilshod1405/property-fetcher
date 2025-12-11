package property

type PFListing struct {
    ID string `json:"id"`

    Title struct {
        En string `json:"en"`
    } `json:"title"`

    Description struct {
        En string `json:"en"`
    } `json:"description"`

    Category       string `json:"category"`
    FurnishingType string `json:"furnishingType"`

    Bathrooms PFIntString `json:"bathrooms"`
    Bedrooms  PFIntString `json:"bedrooms"`

    Size float64 `json:"size"`

    Location struct {
        ID uint `json:"id"`
    } `json:"location"`

    AssignedTo struct {
        ID int64 `json:"id"`
    } `json:"assignedTo"`

    Price struct {
        Amounts struct {
            Sale int64 `json:"sale"`
        } `json:"amounts"`
    } `json:"price"`

    Media struct {
        Images []struct {
            Original struct {
                URL string `json:"url"`
            } `json:"original"`
        } `json:"images"`
    } `json:"media"`

    Reference string `json:"reference"`
}
