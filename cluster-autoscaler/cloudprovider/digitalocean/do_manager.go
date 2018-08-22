/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package digitalocean

import (
	"github.com/digitalocean/godo"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"net/http"
	"net/url"
	"context"
	"fmt"
	"regexp"
//"github.com/digitalocean/godo"
)

var (
	tagRegex = regexp.MustCompile("^pool-.*")
)

type DOManager interface {
	GetNodeGroups() []cloudprovider.NodeGroup
	SetURL(string)
}

type doManagerImpl struct {
	doClient *godo.Client
}

func NewDOManager() DOManager {
	client := godo.NewClient(&http.Client{})

	return &doManagerImpl{
		doClient: client,
	}
}

func (d *doManagerImpl) SetURL(toUrl string) {
	parsed, _ := url.Parse(toUrl)
	d.doClient.BaseURL = parsed
}

func (d *doManagerImpl) GetNodeGroups() []cloudprovider.NodeGroup {
	ctx := context.TODO()

	tags, _, err := d.doClient.Tags.List(ctx, &godo.ListOptions{})
	if err != nil {
		panic(err)
	}

	pools := []cloudprovider.NodeGroup{}

	for _, tag := range tags {
		if ! tagRegex.MatchString(tag.Name) {
			continue
		}

		pools = append(pools, &doPool{
			doManager: d,
			id: tag.Name,
		})
	}

	fmt.Println(tags)
	return pools
}
