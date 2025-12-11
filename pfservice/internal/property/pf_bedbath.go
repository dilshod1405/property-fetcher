package property

import "strconv"

type PFIntString struct {
	Value int
}

func (p *PFIntString) UnmarshalJSON(b []byte) error {
	s := string(b)

	// Remove quotes
	if len(s) >= 2 && s[0] == '"' {
		s = s[1 : len(s)-1]
	}

	if s == "studio" || s == "" {
		p.Value = 0
		return nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		p.Value = 0
		return nil
	}

	p.Value = i
	return nil
}
