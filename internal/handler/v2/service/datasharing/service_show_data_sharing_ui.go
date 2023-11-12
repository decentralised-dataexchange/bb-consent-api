package datasharing

import (
	"html/template"
	"net/http"
	"net/url"
)

func parseQueryParams(r *http.Request) (dataAgreementId string, accessToken string, apiKey string, individualId string, thirdPartyOrgName string, thirdPartyOrgLogoImageUrl string, dataSharingUiRedirectUrl string, authorisationCode string, authorisationRedirectUrl string) {
	query := r.URL.Query()

	dataAgreementId = getStringQueryParam(query, "dataAgreementId")
	accessToken = getStringQueryParam(query, "accessToken")
	apiKey = getStringQueryParam(query, "apiKey")
	individualId = getStringQueryParam(query, "individualId")
	thirdPartyOrgName = getStringQueryParam(query, "thirdPartyOrgName")
	thirdPartyOrgLogoImageUrl = getStringQueryParam(query, "thirdPartyOrgLogoImageUrl")
	dataSharingUiRedirectUrl = getStringQueryParam(query, "dataSharingUiRedirectUrl")
	authorisationCode = getStringQueryParam(query, "authorisationCode")
	authorisationRedirectUrl = getStringQueryParam(query, "authorisationRedirectUrl")

	return dataAgreementId, accessToken, apiKey, individualId, thirdPartyOrgName, thirdPartyOrgLogoImageUrl, dataSharingUiRedirectUrl, authorisationCode, authorisationRedirectUrl
}

func getStringQueryParam(query url.Values, param string) string {
	values, ok := query[param]
	if ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

func ServiceShowDataSharingUiHandler(w http.ResponseWriter, r *http.Request) {
	baseUrl := "https://" + r.Host + "/v2"
	dataAgreementId, accessToken, apiKey, individualId, thirdPartyOrgName, thirdPartyOrgLogoImageUrl, dataSharingUiRedirectUrl, authorisationCode, authorisationRedirectUrl := parseQueryParams(r)
	// HTML template
	templateContent := `
	<!DOCTYPE html>
	<html lang="en">
	  <head>
		<meta charset="UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<title>Consent BB Data Sharing UI</title>
		<link
		  rel="stylesheet"
		  href="https://cdn.jsdelivr.net/gh/decentralised-dataexchange/bb-consent-data-sharing-ui@2023.11.3/dist/consentBbDataSharingUi.css"
		/>
	  </head>
	  <body style="margin: 0px">
		<div id="consentBbDataSharingUi"></div>
	
		<script
		  data-element-id="consentBbDataSharingUi"
		  id="consentBbDataSharingUi-script"
		  src="https://cdn.jsdelivr.net/gh/decentralised-dataexchange/bb-consent-data-sharing-ui@2023.11.3/dist/consentBbDataSharingUi.js"
		></script>
		<script>
			window.ConsentBbDataSharingUi({
				baseUrl: {{ if .BaseUrl }} {{.BaseUrl}} {{ else }} undefined {{ end }},
				dataAgreementId: {{ if .DataAgreementId }} {{.DataAgreementId}} {{ else }} undefined {{ end }},
				accessToken: {{ if .AccessToken }} {{.AccessToken}} {{ else }} undefined {{ end }},
				apiKey: {{ if .ApiKey }} {{.ApiKey}} {{ else }} undefined {{ end }},
				individualId: {{ if .IndividualId }} {{.IndividualId}} {{ else }} undefined {{ end }},
				thirdPartyOrgName: {{ if .ThirdPartyOrgName }} {{.ThirdPartyOrgName}} {{ else }} undefined {{ end }},
				thirdPartyOrgLogoImageUrl: {{ if .ThirdPartyOrgLogoImageUrl }} {{.ThirdPartyOrgLogoImageUrl}} {{ else }} undefined {{ end }},
				dataSharingUiRedirectUrl: {{ if .DataSharingUiRedirectUrl }} {{.DataSharingUiRedirectUrl}} {{ else }} undefined {{ end }},
				authorisationCode: {{ if .AuthorisationCode }} {{.AuthorisationCode}} {{ else }} undefined {{ end }},
				authorisationRedirectUrl: {{ if .AuthorisationRedirectUrl }} {{.AuthorisationRedirectUrl}} {{ else }} undefined {{ end }},
			});
		</script>
	  </body>
	</html>
	`

	// Create a map to hold dynamic values for template rendering
	data := map[string]interface{}{
		"BaseUrl":                   baseUrl,
		"DataAgreementId":           dataAgreementId,
		"AccessToken":               accessToken,
		"ApiKey":                    apiKey,
		"IndividualId":              individualId,
		"ThirdPartyOrgName":         thirdPartyOrgName,
		"ThirdPartyOrgLogoImageUrl": thirdPartyOrgLogoImageUrl,
		"DataSharingUiRedirectUrl":  dataSharingUiRedirectUrl,
		"AuthorisationCode":         authorisationCode,
		"AuthorisationRedirectUrl":  authorisationRedirectUrl,
	}

	// Parse the HTML template
	tmpl := template.New("my-template")
	tmpl, err := tmpl.Parse(templateContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template with the data
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
