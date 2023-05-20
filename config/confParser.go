package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shadiestgoat/log"
)

type caseInsensitiveInclusion map[string]bool

func panicLoad(key, section, err string) {
	panic("While setting a value for " + keyFormat(section, key) + ": " + err + "!")
}

type Parsable interface {
	Parse(inp string) error
}

func keyFormat(sec, key string) string {
	return "[" + sec + "].[" + key + "]"
}

func normMod(mod string) string {
	if mod == "" {
		return "GENERAL"
	}

	return strings.ToUpper(mod)
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
		"Games":     &Games,
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

	moduleWarnedKeys := map[string]map[string][]string{}
	moduleRequiredKeys := map[string][]string{}

	addToWarnings := func(mod, sec, key, consequence string) {
		mod = normMod(mod)

		if moduleWarnedKeys[mod] == nil {
			moduleWarnedKeys[mod] = map[string][]string{}
		}

		if moduleWarnedKeys[mod][consequence] == nil {
			moduleWarnedKeys[mod][consequence] = []string{}
		}

		moduleWarnedKeys[mod][consequence] = append(moduleWarnedKeys[mod][consequence], keyFormat(sec, key))
	}

	addToRequired := func(mod, sec, key string) {
		mod = normMod(mod)
		if moduleRequiredKeys[mod] == nil {
			moduleRequiredKeys[mod] = []string{}
		}

		moduleRequiredKeys[mod] = append(moduleRequiredKeys[mod], keyFormat(sec, key))
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

			spl := strings.SplitN(tagV, ",", 4)

			var (
				key         = spl[0]
				isRequired  = false
				module      = ""
				consequence = ""
			)

			if len(spl) >= 2 {
				module = spl[1]

				if len(spl) == 4 {
					isRequired = spl[2] == "required"
					consequence = spl[3]
				} else if len(spl) == 3 {
					if spl[2] == "required" {
						isRequired = true
					} else if spl[2] != "" {
						consequence = spl[2]
					}
				}
			}

			givenVal := conf[key]

			if givenVal == "" {
				if len(spl) == 2 {
					if isRequired {
						addToRequired(module, sec, key)
					} else if consequence != "" {
						addToWarnings(module, sec, key, consequence)
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
				case map[string]bool:
					tmpM := map[string]bool{}

					spl := strings.Split(givenVal, " ")

					for _, s := range spl {
						tmpM[s] = true
					}

					vToSet = tmpM
				case caseInsensitiveInclusion:
					tmpM := map[string]bool{}

					spl := strings.Split(givenVal, " ")

					for _, s := range spl {
						tmpM[strings.ToLower(s)] = true
					}

					vToSet = tmpM
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

	for v := range General.Disabled {
		v = strings.ToLower(v)

		delete(moduleRequiredKeys, v)
		delete(moduleWarnedKeys, v)
	}

	if len(moduleRequiredKeys) != 0 {
		// we will not being any further than this module, so its ok to init the log right now!
		initLog()
	}

	for m, consequenceMap := range moduleWarnedKeys {
		for con, keys := range consequenceMap {
			if con == "" || len(keys) == 0 {
				continue
			}

			if len(keys) == 1 {
				log.Warn("%s: Due to %s not being set, %s", m, keys[0], con)
			} else {
				keysStr := "- " + strings.Join(keys, "\n- ")

				log.Warn("%s: Due to the following keys not being set, %s:\n%s", m, con, keysStr)
			}
		}
	}

	if len(moduleWarnedKeys) != 0 {
		log.PrintDebug("To avoid the above warnings, either set those keys or disable the associated module using the DISABLE config key")
		log.PrintDebug("For more info, read the config.template.conf file!")
	}

	time.Sleep(150 * time.Millisecond)

	if len(moduleRequiredKeys) != 0 {
		str := "There are required keys that are not set! To fix this, either set these keys or disable the modules associated with them!"

		for m, keys := range moduleRequiredKeys {
			str += "\n" + m + ":"

			for _, k := range keys {
				str += "\n- " + k
			}
		}

		log.Fatal(str)
	}
}
