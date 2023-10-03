package handlerv1

import (
	"time"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/consent"
	"github.com/bb-consent/api/src/org"
)

func getConsentsWithPurpose(purposeID string, c consent.Consents) []consent.Consent {
	for _, p := range c.Purposes {
		if p.ID == purposeID {
			return p.Consents
		}
	}
	return []consent.Consent{}
}

func getConsentWithTemplateID(templateID string, consents []consent.Consent) consent.Consent {
	for _, c := range consents {
		if c.TemplateID == templateID {
			return c
		}
	}
	return consent.Consent{}
}

func getConsentCount(cp consentsAndPurpose) ConsentCount {
	var c ConsentCount
	var disallowCount = 0

	for _, p := range cp.Consents {
		c.Total++
		if (p.Status.Consented == common.ConsentStatusDisAllow) || (p.Status.Consented == "DisAllow") {
			disallowCount++
		}
	}
	c.Consented = c.Total - disallowCount
	return c
}

func createConsentResponse(templates []org.Template, consents []consent.Consent, purpose org.Purpose) consentsAndPurpose {
	var cp consentsAndPurpose
	for _, template := range templates {
		var conResp consentResp
		conResp.ID = template.ID
		conResp.Description = template.Consent

		// Fetching consents matching a Template ID
		c := getConsentWithTemplateID(template.ID, consents)

		if (consent.Consent{}) == c {
			if purpose.LawfulUsage {
				conResp.Status.Consented = common.ConsentStatusAllow
			} else {
				conResp.Status.Consented = common.ConsentStatusDisAllow
			}
		} else {
			conResp.Status.Consented = c.Status.Consented
			conResp.Value = c.Value
			if c.Status.Days != 0 {
				conResp.Status.Days = c.Status.Days
				conResp.Status.Remaining = c.Status.Days - int((time.Now().Sub(c.Status.TimeStamp).Hours())/24)
				if conResp.Status.Remaining <= 0 {
					conResp.Status.Consented = common.ConsentStatusDisAllow
					conResp.Status.Remaining = 0
				} else {
					conResp.Status.TimeStamp = c.Status.TimeStamp
				}

			}
		}
		cp.Consents = append(cp.Consents, conResp)
	}
	cp.Purpose = purpose
	cp.Count = getConsentCount(cp)
	return cp
}

func createConsentGetResponse(c consent.Consents, o org.Organization) ConsentsResp {
	var cResp ConsentsResp
	cResp.ID = c.ID.Hex()
	cResp.OrgID = c.OrgID
	cResp.UserID = c.UserID

	for _, p := range o.Purposes {
		// Filtering templates corresponding to the purpose ID
		templatesWithPurpose := getTemplateswithPurpose(p.ID, o.Templates)

		// Filtering consents corresponding to purpose ID
		cons := getConsentsWithPurpose(p.ID, c)

		conResp := createConsentResponse(templatesWithPurpose, cons, p)

		cResp.ConsentsAndPurposes = append(cResp.ConsentsAndPurposes, conResp)
	}
	return cResp
}

func getTemplateswithPurpose(purposeID string, templates []org.Template) []org.Template {
	var t []org.Template
	for _, template := range templates {
		for _, pID := range template.PurposeIDs {
			if pID == purposeID {
				t = append(t, template)
				break
			}
		}
	}
	return t
}

func getPurposeFromID(p []org.Purpose, purposeID string) org.Purpose {
	for _, e := range p {
		if e.ID == purposeID {
			return e
		}
	}
	return org.Purpose{}
}
