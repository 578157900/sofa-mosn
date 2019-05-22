package cluster

import (
	"sofastack.io/sofa-mosn/pkg/metrics"
	"sofastack.io/sofa-mosn/pkg/types"
)

func newHostStats(clustername string, addr string) types.HostStats {
	s := metrics.NewHostStats(clustername, addr)

	return types.HostStats{
		UpstreamConnectionTotal:                        s.Counter(metrics.UpstreamConnectionTotal),
		UpstreamConnectionClose:                        s.Counter(metrics.UpstreamConnectionClose),
		UpstreamConnectionActive:                       s.Counter(metrics.UpstreamConnectionActive),
		UpstreamConnectionConFail:                      s.Counter(metrics.UpstreamConnectionConFail),
		UpstreamConnectionLocalClose:                   s.Counter(metrics.UpstreamConnectionLocalClose),
		UpstreamConnectionRemoteClose:                  s.Counter(metrics.UpstreamConnectionRemoteClose),
		UpstreamConnectionLocalCloseWithActiveRequest:  s.Counter(metrics.UpstreamConnectionLocalCloseWithActiveRequest),
		UpstreamConnectionRemoteCloseWithActiveRequest: s.Counter(metrics.UpstreamConnectionRemoteCloseWithActiveRequest),
		UpstreamConnectionCloseNotify:                  s.Counter(metrics.UpstreamConnectionCloseNotify),
		UpstreamRequestTotal:                           s.Counter(metrics.UpstreamRequestTotal),
		UpstreamRequestActive:                          s.Counter(metrics.UpstreamRequestActive),
		UpstreamRequestLocalReset:                      s.Counter(metrics.UpstreamRequestLocalReset),
		UpstreamRequestRemoteReset:                     s.Counter(metrics.UpstreamRequestRemoteReset),
		UpstreamRequestTimeout:                         s.Counter(metrics.UpstreamRequestTimeout),
		UpstreamRequestFailureEject:                    s.Counter(metrics.UpstreamRequestFailureEject),
		UpstreamRequestPendingOverflow:                 s.Counter(metrics.UpstreamRequestPendingOverflow),
		UpstreamRequestDuration:                        s.Histogram(metrics.UpstreamRequestDuration),
		UpstreamRequestDurationTotal:                   s.Counter(metrics.UpstreamRequestDurationTotal),
		UpstreamResponseSuccess:                        s.Counter(metrics.UpstreamResponseSuccess),
		UpstreamResponseFailed:                         s.Counter(metrics.UpstreamResponseFailed),
	}
}

func newClusterStats(clustername string) types.ClusterStats {
	s := metrics.NewClusterStats(clustername)
	return types.ClusterStats{
		UpstreamConnectionTotal:                        s.Counter(metrics.UpstreamConnectionTotal),
		UpstreamConnectionClose:                        s.Counter(metrics.UpstreamConnectionClose),
		UpstreamConnectionActive:                       s.Counter(metrics.UpstreamConnectionActive),
		UpstreamConnectionConFail:                      s.Counter(metrics.UpstreamConnectionConFail),
		UpstreamConnectionRetry:                        s.Counter(metrics.UpstreamConnectionRetry),
		UpstreamConnectionLocalClose:                   s.Counter(metrics.UpstreamConnectionLocalClose),
		UpstreamConnectionRemoteClose:                  s.Counter(metrics.UpstreamConnectionRemoteClose),
		UpstreamConnectionLocalCloseWithActiveRequest:  s.Counter(metrics.UpstreamConnectionLocalCloseWithActiveRequest),
		UpstreamConnectionRemoteCloseWithActiveRequest: s.Counter(metrics.UpstreamConnectionRemoteCloseWithActiveRequest),
		UpstreamConnectionCloseNotify:                  s.Counter(metrics.UpstreamConnectionCloseNotify),
		UpstreamBytesReadTotal:                         s.Counter(metrics.UpstreamBytesReadTotal),
		UpstreamBytesWriteTotal:                        s.Counter(metrics.UpstreamBytesWriteTotal),
		UpstreamRequestTotal:                           s.Counter(metrics.UpstreamRequestTotal),
		UpstreamRequestActive:                          s.Counter(metrics.UpstreamRequestActive),
		UpstreamRequestLocalReset:                      s.Counter(metrics.UpstreamRequestLocalReset),
		UpstreamRequestRemoteReset:                     s.Counter(metrics.UpstreamRequestRemoteReset),
		UpstreamRequestRetry:                           s.Counter(metrics.UpstreamRequestRetry),
		UpstreamRequestRetryOverflow:                   s.Counter(metrics.UpstreamRequestRetryOverflow),
		UpstreamRequestTimeout:                         s.Counter(metrics.UpstreamRequestTimeout),
		UpstreamRequestFailureEject:                    s.Counter(metrics.UpstreamRequestFailureEject),
		UpstreamRequestPendingOverflow:                 s.Counter(metrics.UpstreamRequestPendingOverflow),
		UpstreamRequestDuration:                        s.Histogram(metrics.UpstreamRequestDuration),
		UpstreamRequestDurationTotal:                   s.Counter(metrics.UpstreamRequestDurationTotal),
		UpstreamResponseSuccess:                        s.Counter(metrics.UpstreamResponseSuccess),
		UpstreamResponseFailed:                         s.Counter(metrics.UpstreamResponseFailed),
		LBSubSetsFallBack:                              s.Counter(metrics.UpstreamLBSubSetsFallBack),
		LBSubSetsActive:                                s.Counter(metrics.UpstreamLBSubSetsActive),
		LBSubsetsCreated:                               s.Counter(metrics.UpstreamLBSubsetsCreated),
		LBSubsetsRemoved:                               s.Counter(metrics.UpstreamLBSubsetsRemoved),
	}
}
