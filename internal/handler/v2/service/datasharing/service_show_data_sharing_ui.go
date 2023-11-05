package datasharing

import (
	"html/template"
	"net/http"
)

func ServiceShowDataSharingUiHandler(w http.ResponseWriter, r *http.Request) {
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
		  href="https://cdn.jsdelivr.net/gh/decentralised-dataexchange/bb-consent-data-sharing-ui/dist/consentBbDataSharingUi.css"
		/>
	  </head>
	  <body style="margin: 0px">
		<div id="consentBbDataSharingUi"></div>
	
		<script
		  data-element-id="consentBbDataSharingUi"
		  id="consentBbDataSharingUi-script"
		  src="https://cdn.jsdelivr.net/gh/decentralised-dataexchange/bb-consent-data-sharing-ui/dist/consentBbDataSharingUi.js"
		></script>
		<script>
		  window.ConsentBbDataSharingUi();
		</script>
	  </body>
	</html>
	`

	// Parse the HTML template
	tmpl := template.New("my-template")
	tmpl, err := tmpl.Parse(templateContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template with the data
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
