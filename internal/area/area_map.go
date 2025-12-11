package area

var PF_TO_DJANGO_AREA = map[uint]uint{
    3782: 1,
    1001: 2,
    1002: 3,
}


func MapPFToDjangoArea(pfAreaID uint) uint {
    if v, ok := PF_TO_DJANGO_AREA[pfAreaID]; ok {
        return v
    }
    return 1
}
