package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/utils"
)

var consequenceSuffix = []string{
	" <3",
	"!",
	" :(",
}

func panicLoad(key, section, err string) {
	panic("While setting a value for '" + section + "." + key + "': " + err + "!")
}

type Parsable interface {
	Parse(inp string) error
}

func load() {
	secMap := map[string]any{
		"Debug":     &debugV,
		"General":   &General,
		"Discord":   &Discord,
		"Channels":  &Channels,
		"Warnings":  &Warnings,
		"Donations": &Donations,
		"Twitch":    &Twitch,
		"XP":        &XP,
	}

	b1, _ := os.ReadFile("config.conf")
	b2, _ := os.ReadFile("secrets.conf")

	fullFile := string(b1) + "\n" + string(b2)

	rawConfig := map[string]map[string]string{}

	lines := strings.Split(fullFile, "\n")

	tmpValues := map[string]string{}
	curSection := ""

	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if l[0] == ';' {
			continue
		}
		if l[0] == '[' && (len(l) < 3 || l[len(l)-1] != ']') {
			log.Warn("Line '%s' is not valid, since it starts out as a heading, but doesn't conclude as such! Ignoring...", l)
			continue
		}

		if l[0] == '[' {
			if curSection != "" {
				if rawConfig[curSection] != nil {
					for k, v := range tmpValues {
						rawConfig[curSection][k] = v
					}
				} else {
					rawConfig[curSection] = tmpValues
				}
			}

			curSection = strings.TrimSpace(l[1 : len(l)-1])

			if secMap[curSection] == nil {
				log.Warn("Section '%s' not recognized, ignoring it's values...", curSection)
				curSection = ""
			}

			tmpValues = map[string]string{}

			continue
		}

		if curSection == "" {
			continue
		}

		v := strings.SplitN(l, "=", 2)

		if len(v) == 2 {
			v[0] = strings.ToLower(strings.TrimSpace(v[0]))

			v[1] = strings.TrimSpace(v[1])

			if len(v[1]) >= 2 && v[1][0] == '"' && v[1][len(v[1])-1] == '"' {
				v[1] = v[1][1 : len(v[1])-1]
			}
		} else {
			continue
		}

		if v[0] == "" || v[1] == "" {
			continue
		}

		tmpValues[v[0]] = v[1]
	}

	if curSection != "" {
		if rawConfig[curSection] != nil {
			for k, v := range tmpValues {
				rawConfig[curSection][k] = v
			}
		} else {
			rawConfig[curSection] = tmpValues
		}
	}

	for sec, p := range secMap {
		t := reflect.TypeOf(p)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		sV := reflect.Indirect(reflect.ValueOf(p))

		conf := rawConfig[sec]

		if conf == nil {
			conf = map[string]string{}
		}

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tagV := f.Tag.Get("conf")
			if tagV == "" {
				continue
			}

			spl := strings.SplitN(tagV, ",", 2)
			key := spl[0]
			givenVal := conf[key]

			if givenVal == "" {
				if len(spl) == 2 {
					if spl[1] == "required" {
						panic(sec + "." + key + " is required! For more info look at the example conf!")
					} else {
						log.Warn("Value '%s' under section '%s' is not set, so %s%s", key, sec, spl[1], consequenceSuffix[utils.RandInt(0, len(consequenceSuffix)-1)])
					}
				}
			} else {
				vF := sV.Field(i)
				vI := vF.Interface()

				var vToSet any

				switch vI.(type) {
				case string:
					vToSet = givenVal
				case int:
					v, err := strconv.Atoi(givenVal)
					if err != nil {
						panicLoad(key, sec, "not a valid int value")
					}
					vToSet = v
				case float64:
					v, err := strconv.ParseFloat(givenVal, 64)
					if err != nil {
						panicLoad(key, sec, "not a valid float value")
					}
					vToSet = v
				case bool:
					givenVal = strings.ToLower(givenVal)
					switch givenVal {
					case "t", "true", "yes", "1":
						vToSet = true
					case "f", "false", "no", "0":
						vToSet = false
					default:
						panicLoad(key, sec, "not a valid bool value")
					}
				case []string:
					vToSet = strings.Split(givenVal, " ")
				case []int:
					v := strings.Split(givenVal, " ")
					vs := []int{}

					for _, tmpV := range v {
						val, err := strconv.Atoi(tmpV)
						if err != nil {
							panicLoad(key, sec, "not a valid []int value")
						}
						vs = append(vs, val)
					}

					vToSet = vs
				default:
					if vI, ok := vI.(Parsable); ok {
						err := vI.Parse(givenVal)
						if err != nil {
							panic("While setting a value for '%s'.'%s', not a valid value!")
						}
					} else {
						panicLoad(key, sec, "unknown type")
					}
				}

				if vToSet != nil {
					rV := reflect.ValueOf(vToSet)
					vF.Set(rV)
				}
			}
		}
	}
}
