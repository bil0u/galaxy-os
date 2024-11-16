package features

func NewFeatureSet(features ...Feature) *FeatureSet {
	fs := FeatureSet{}
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

type FeatureSet map[string]Feature

func (fs FeatureSet) WithKey(key string) *Feature {
	if f, ok := fs[key]; ok {
		return &f
	}
	return nil
}

func (fs FeatureSet) Get(matchFunc func(f Feature) bool) []*Feature {
	features := make([]*Feature, 0)
	for _, f := range fs {
		if matchFunc(f) {
			features = append(features, &f)
		}
	}
	return features
}

func (fs FeatureSet) Bot() []*Feature {
	return fs.Get(func(f Feature) bool {
		return f.Type == BotFeature
	})
}

func (fs FeatureSet) Guild() []*Feature {
	return fs.Get(func(f Feature) bool {
		return f.Type == GuildFeature
	})
}
