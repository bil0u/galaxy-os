package feature

func NewSet(features ...Feature) *Set {
	fs := Set{}
	for _, f := range features {
		if err := f.IsValid(); err != nil {
			return nil
		}
		featureKey := f.Key
		if _, ok := fs[featureKey]; ok {
			continue
		}
		fs[featureKey] = f
	}
	return &fs
}

type Set map[string]Feature

func (fs Set) WithKey(key string) *Feature {
	if f, ok := fs[key]; ok {
		return &f
	}
	return nil
}

func (fs Set) Get(matchFunc func(f Feature) bool) []*Feature {
	features := make([]*Feature, 0)
	for _, f := range fs {
		if matchFunc(f) {
			features = append(features, &f)
		}
	}
	return features
}

func (fs Set) Bot() []*Feature {
	return fs.Get(func(f Feature) bool {
		return f.Type == BotFeature
	})
}

func (fs Set) Guild() []*Feature {
	return fs.Get(func(f Feature) bool {
		return f.Type == GuildFeature
	})
}
