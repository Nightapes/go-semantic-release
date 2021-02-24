package integrations

import (
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/sjson"
	"io/ioutil"
)

func (i *Integrations) updateNPM() error {

	npmConfig := i.config.NPM
	if npmConfig.Path == "" {
		npmConfig.Path = "./package.json"
	}

	log.Debugf("Set version %s to %s", i.version.Next.Version, npmConfig.Path)
	data, err := ioutil.ReadFile(npmConfig.Path)
	if err != nil {
		return err
	}

	newData, err := sjson.Set(string(data), "version", i.version.Next.Version)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(npmConfig.Path, []byte(newData), 0777)
}
