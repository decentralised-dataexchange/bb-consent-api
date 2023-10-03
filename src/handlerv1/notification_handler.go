package handlerv1

import "github.com/bb-consent/api/src/org"

func getTemplateFromOrg(o org.Organization, templateID string) org.Template {
	for _, t := range o.Templates {
		if t.ID == templateID {
			return t
		}
	}
	return org.Template{}
}
