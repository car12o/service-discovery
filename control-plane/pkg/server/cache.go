package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"

	"github.com/car12o/service-discovery-exp/control-plane/pkg/snapshot"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"go.etcd.io/etcd/clientv3"
)

type ServerCacheManager struct {
	NodeId         string
	Cache          cachev3.SnapshotCache
	Etcd           *clientv3.Client
	Log            logger
	snapshotConfig snapshot.Config
}

func NewServerCacheManager(nodeId string, etcd *clientv3.Client, debug bool) *ServerCacheManager {
	log := logger{Debug: debug}
	return &ServerCacheManager{
		NodeId: nodeId,
		Cache:  cachev3.NewSnapshotCache(false, cachev3.IDHash{}, log),
		Etcd:   etcd,
		Log:    log,
		snapshotConfig: snapshot.Config{
			Version: 0,
			Nodes:   map[snapshot.UpstreamHost]*snapshot.Node{},
		},
	}
}

func (scm *ServerCacheManager) GetCache() cachev3.SnapshotCache {
	return scm.Cache
}

func (scm *ServerCacheManager) Start() {
	go scm.watchConfig()
}

func (scm *ServerCacheManager) watchConfig() {
	wChan := scm.Etcd.Watch(context.Background(), "service", clientv3.WithPrefix())
	for res := range wChan {
		for _, event := range res.Events {
			node := new(snapshot.Node)
			dec := gob.NewDecoder(bytes.NewReader(event.Kv.Value))
			if err := dec.Decode(node); err != nil {
				log.Println(err)
				continue
			}
			scm.snapshotConfig.Nodes[snapshot.UpstreamHost(node.UpstreamHost)] = node
		}

		scm.updateCache()
	}
}

func (scm *ServerCacheManager) updateCache() {
	scm.snapshotConfig.Version += 1
	snapshot, err := snapshot.Generate(scm.snapshotConfig)
	if err != nil {
		log.Println(err)
		return
	}
	if err := scm.Cache.SetSnapshot(scm.NodeId, *snapshot); err != nil {
		log.Println(err)
		return
	}
	scm.Log.Debugf("will serve snapshot %+v", snapshot)
}
