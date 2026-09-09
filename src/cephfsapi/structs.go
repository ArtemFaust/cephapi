package cephfsapi

// Структура описывающая доступные подключения (rados)
type Connection struct {
	RADOS RADOS `yaml:"RADOS"`
}

type RADOS struct {
	EndPoints []struct {
		Name                 string `yaml:"Name"`
		Fsid                 string `yaml:"Fsid"`
		MonHosts             string `yaml:"MonHosts"`
		DefaultMetaCrushRule string `yaml:"DefaultMetaCrushRule"`
		DefaultDataCrushRule string `yaml:"DefaultDataCrushRule"`
		KeyRing              string `yaml:"KeyRing"`
	} `yaml:"EndPoints"`
}

// Структура описывающая информацию о пуле
type PoolInfo struct {
	PoolID                            int64               `json:"pool_id"`
	PoolName                          string              `json:"pool_name"`
	CreateTime                        string              `json:"create_time"`
	Flags                             int64               `json:"flags"`
	FlagsNames                        string              `json:"flags_names"`
	Type                              int64               `json:"type"`
	Size                              int64               `json:"size"`
	MinSize                           int64               `json:"min_size"`
	CrushRule                         int64               `json:"crush_rule"`
	PeeringCrushBucketCount           int64               `json:"peering_crush_bucket_count"`
	PeeringCrushBucketTarget          int64               `json:"peering_crush_bucket_target"`
	PeeringCrushBucketBarrier         int64               `json:"peering_crush_bucket_barrier"`
	PeeringCrushBucketMandatoryMember int64               `json:"peering_crush_bucket_mandatory_member"`
	IsStretchPool                     bool                `json:"is_stretch_pool"`
	ObjectHash                        int64               `json:"object_hash"`
	PGAutoscaleMode                   string              `json:"pg_autoscale_mode"`
	PGNum                             int64               `json:"pg_num"`
	PGPlacementNum                    int64               `json:"pg_placement_num"`
	PGPlacementNumTarget              int64               `json:"pg_placement_num_target"`
	PGNumTarget                       int64               `json:"pg_num_target"`
	PGNumPending                      int64               `json:"pg_num_pending"`
	LastPGMergeMeta                   LastPGMergeMeta     `json:"last_pg_merge_meta"`
	LastChange                        string              `json:"last_change"`
	LastForceOpResend                 string              `json:"last_force_op_resend"`
	LastForceOpResendPrenautilus      string              `json:"last_force_op_resend_prenautilus"`
	LastForceOpResendPreluminous      string              `json:"last_force_op_resend_preluminous"`
	Auid                              int64               `json:"auid"`
	SnapMode                          string              `json:"snap_mode"`
	SnapSeq                           int64               `json:"snap_seq"`
	SnapEpoch                         int64               `json:"snap_epoch"`
	PoolSnaps                         []interface{}       `json:"pool_snaps"`
	RemovedSnaps                      string              `json:"removed_snaps"`
	QuotaMaxBytes                     int64               `json:"quota_max_bytes"`
	QuotaMaxObjects                   int64               `json:"quota_max_objects"`
	Tiers                             []interface{}       `json:"tiers"`
	TierOf                            int64               `json:"tier_of"`
	ReadTier                          int64               `json:"read_tier"`
	WriteTier                         int64               `json:"write_tier"`
	CacheMode                         string              `json:"cache_mode"`
	TargetMaxBytes                    int64               `json:"target_max_bytes"`
	TargetMaxObjects                  int64               `json:"target_max_objects"`
	CacheTargetDirtyRatioMicro        int64               `json:"cache_target_dirty_ratio_micro"`
	CacheTargetDirtyHighRatioMicro    int64               `json:"cache_target_dirty_high_ratio_micro"`
	CacheTargetFullRatioMicro         int64               `json:"cache_target_full_ratio_micro"`
	CacheMinFlushAge                  int64               `json:"cache_min_flush_age"`
	CacheMinEvictAge                  int64               `json:"cache_min_evict_age"`
	ErasureCodeProfile                string              `json:"erasure_code_profile"`
	HitSetParams                      HitSetParams        `json:"hit_set_params"`
	HitSetPeriod                      int64               `json:"hit_set_period"`
	HitSetCount                       int64               `json:"hit_set_count"`
	UseGmtHitset                      bool                `json:"use_gmt_hitset"`
	MinReadRecencyForPromote          int64               `json:"min_read_recency_for_promote"`
	MinWriteRecencyForPromote         int64               `json:"min_write_recency_for_promote"`
	HitSetGradeDecayRate              int64               `json:"hit_set_grade_decay_rate"`
	HitSetSearchLastN                 int64               `json:"hit_set_search_last_n"`
	GradeTable                        []interface{}       `json:"grade_table"`
	StripeWidth                       int64               `json:"stripe_width"`
	ExpectedNumObjects                int64               `json:"expected_num_objects"`
	FastRead                          bool                `json:"fast_read"`
	NonprimaryShards                  string              `json:"nonprimary_shards"`
	Options                           Options             `json:"options"`
	ApplicationMetadata               ApplicationMetadata `json:"application_metadata"`
	ReadBalance                       ReadBalance         `json:"read_balance"`
	Subvolumes                        []Subvolume
}

type ApplicationMetadata struct {
	Cephfs Cephfs `json:"cephfs"`
}

type Cephfs struct {
	Data     string `json:"data"`
	MetaData string `json:"metadata"`
}

type HitSetParams struct {
	Type string `json:"type"`
}

type LastPGMergeMeta struct {
	SourcePgid       string `json:"source_pgid"`
	ReadyEpoch       int64  `json:"ready_epoch"`
	LastEpochStarted int64  `json:"last_epoch_started"`
	LastEpochClean   int64  `json:"last_epoch_clean"`
	SourceVersion    string `json:"source_version"`
	TargetVersion    string `json:"target_version"`
}

type Options struct {
	CompressionAlgorithm string `json:"compression_algorithm"`
}

type ReadBalance struct {
	ScoreType                      string  `json:"score_type"`
	ScoreActing                    float64 `json:"score_acting"`
	ScoreStable                    float64 `json:"score_stable"`
	OptimalScore                   float64 `json:"optimal_score"`
	RawScoreActing                 float64 `json:"raw_score_acting"`
	RawScoreStable                 float64 `json:"raw_score_stable"`
	PrimaryAffinityWeighted        float64 `json:"primary_affinity_weighted"`
	AveragePrimaryAffinity         float64 `json:"average_primary_affinity"`
	AveragePrimaryAffinityWeighted float64 `json:"average_primary_affinity_weighted"`
}

type Subvolume struct {
	Name string `json:"name"`
	Info SubvolumeInfo
}

type SubvolumeInfo struct {
	Atime         string   `json:"atime"`
	BytesPcent    string   `json:"bytes_pcent"`
	BytesQuota    int64    `json:"bytes_quota"`
	BytesUsed     int64    `json:"bytes_used"`
	Casesensitive bool     `json:"casesensitive"`
	CreatedAt     string   `json:"created_at"`
	Ctime         string   `json:"ctime"`
	DataPool      string   `json:"data_pool"`
	Earmark       string   `json:"earmark"`
	Features      []string `json:"features"`
	Flavor        int64    `json:"flavor"`
	Gid           int64    `json:"gid"`
	Mode          int64    `json:"mode"`
	MonAddrs      []string `json:"mon_addrs"`
	Mtime         string   `json:"mtime"`
	Normalization string   `json:"normalization"`
	Path          string   `json:"path"`
	PoolNamespace string   `json:"pool_namespace"`
	State         string   `json:"state"`
	Type          string   `json:"type"`
	Uid           int64    `json:"uid"`
}
