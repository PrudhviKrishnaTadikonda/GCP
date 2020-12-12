package main

import (
	"context"
//	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"golang.org/x/oauth2/google"
	iam "google.golang.org/api/iam/v1"
)

var (
	// iamService is a client for calling the signBlob API.
	iamService *iam.Service

	// serviceAccountName represents Service Account Name.
	// See more details: https://cloud.google.com/iam/docs/service-accounts
	serviceAccountName string

	// serviceAccountID follows the below format.
	// "projects/%s/serviceAccounts/%s"
	serviceAccountID string

	// uploadableBucket is the destination bucket.
	// All users will upload files directly to this bucket by using generated Signed URL.
	uploadableBucket string
)

func signHandler(w http.ResponseWriter, r *http.Request) {
	// Accepts only POST method.
	// Otherwise, this handler returns 405.
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ct := r.FormValue("content_type")
	if ct == "" {
		http.Error(w, "content_type must be set", http.StatusBadRequest)
		return
	}

	// Generates an object key for use in new Cloud Storage Object.
	// It's not duplicate with any object keys because of UUID.
	key := uuid.New().String()
	if ext := r.FormValue("ext"); ext != "" {
		key += fmt.Sprintf(".%s", ext)
	}

	// Generates a signed URL for use in the PUT request to GCS.
	// Generated URL should be expired after 15 mins.
	url, err := storage.SignedURL(uploadableBucket, key, &storage.SignedURLOptions{
		GoogleAccessID: serviceAccountName,
		Method:         "PUT",
		Expires:        time.Now().Add(15 * time.Minute),
		ContentType:    ct,
		// To avoid management for private key, use SignBytes instead of PrivateKey.
		// In this example, we are using the `iam.serviceAccounts.signBlob` API for signing bytes.
		// If you hope to avoid API call for signing bytes every time,
		// you can use self hosted private key and pass it in Privatekey.
		//SignBytes: func(b []byte) ([]byte, error) {
			//resp, err := iamService.Projects.ServiceAccounts.SignBlob(
				//serviceAccountID,
			//	&iam.SignBlobRequest{BytesToSign: base64.StdEncoding.EncodeToString(b)},
			//).Context(r.Context()).Do()
			//if err != nil {
		//		return nil, err
			//}
			//return base64.StdEncoding.DecodeString(resp.Signature)
		//},
		
		PrivateKey: []byte("MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCftHhiE+sydLnQJEphHNISbJ+sEVRSd/faeSb2jfd46LwpBwZD0BDkqlkE326y5BfBS0hkN7TnXjjAGEnBonABdwjMZ5nrUzar5YzrlTu8hsf1NhDgegu7xe1P4vdTkJggPceiIonUwJ7iehC+auzZNYuNvt4aUkclkW5bRo2LzpNylyPukIzqPAkox1C+OazOfpCoqgA7eF/9Q+Eql93Qf14hwCgMEB3BwDJKxxAZ3XABr6Y04jUnmnPpVXmlESh+H5cX7vcR8PqClGsQCS57Kg9HQsGZCb2FfJ8zJyg2tSbMTCQUEFKIgKB4IHKeBewYmLPUw/tiBPmTzCWvaK8XAgMBAAECggEADFEQeOcc9cSCl6ns1V9+vggD9z+6fcYdzQrlG94bO1QFPhhgfdApAis3EVHWYwdSfUhI892t+HgDFW5Qf3v/el8LqsP33wO8ZpAEb1JdJVWAGPB+zbXr4lZJ6CYF0ztTPo1OygZmFD/eVD+vIeUm96HfB0ma+tmxZGIHNxSr+IYQ4+puGYZ/JultHoqSy4YPt8uXVzt5lDALbKdkDs5MF6mQ59rwa4eUi5LQETG3Ryp3ZSySMi0Be1VCq+gwUHWk9gJLvXnEf0e0Ugl4UJ/v/RoHgwoxOz/KegmNG/gSBLLGRifgUsTo0OrC/N5H8+RDHBixrCrNyyPXxUXulgRJoQKBgQDMGr2RqEPpZBkuupexzQdJACF+3WR5PfGycWjZIgadVWPedr1aidXfT4ZbZ9p7Y5EVnlJomRid74JTCZEwchzQ9V4xySewJArbpMi+msNuoFg/TnDEBAM2O/3XbnGsQrYzPN5AsYmcpfQzpamxTHxlft0tYr3ZtWI7/BeI+2DPtwKBgQDIT76vhF79ieKl1lACeXpPiHs6zyVAqi97wesnFQEU/BmHsJxNwYUI79cE8SrylSKvcpYTk6xaRm+pUlV3xOtn/29+Br5iYlu1nmhd4+JbvkCJnFFtYTBwoo/OdppJKwHVtKY7dnsKnveqfKY2jJbXceqZzkdpQsFWcsBN39BboQKBgQCOwuV1zEw0I1+536nbI53E4eKL6i8s3rcAKXM87R/TTLbeFA++FEsUN3uy06FuTOZeSK87mlotnil6C2cSi768KeQIzrqD6bHukAQZzgaEioMvRJ57fJMCjFOxK/82jjMDA8AxX/zxJOL6fRWEfgtEssfhxv8kGErtyhZsKeg9YQKBgGlULWOjikNtZsVnHOlAMUVy8cFpvR/0nUVJIbqKO+hp647DGl10ndymKP1LRxcJvpRc/3dJ1n4dvYdeaNyyqkokMd8l8qRPLgQhSKXeN1+gedUiYlrOmScRA+c/zD8fIzbZZ/OqiGZ8UqTOKKRUZtjg6Mh5hGlgFcO8UUxhnPEhAoGAGNDF3nXeKo6ekZgE0dJzHtU8JnGsDq6A7kfYDa9MdfYI510XQOaAwChsusiiJ89eQGpSgoRwvwspgoBcPRH1lS6fW1nbqTZZVLf2HBbej1vy/x721ezr0ELSOWG3EFjhpJyY/rNJaco9/swJYgiPRE7lGLTiIoTVkcW4A5Hj/bk="), 
	})
	if err != nil {
		log.Printf("sign: failed to sign, err = %v\n", err)
		http.Error(w, "failed to sign by internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, url)
}

func main() {
	cred, err := google.DefaultClient(context.Background(), iam.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}
	iamService, err = iam.New(cred)
	if err != nil {
		log.Fatal(err)
	}

	uploadableBucket = os.Getenv("UPLOADABLE_BUCKET")
	serviceAccountName = os.Getenv("SERVICE_ACCOUNT")
	serviceAccountID = fmt.Sprintf(
		"projects/%s/serviceAccounts/%s",
		os.Getenv("GOOGLE_CLOUD_PROJECT"),
		serviceAccountName,
	)

	http.HandleFunc("/sign", signHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
