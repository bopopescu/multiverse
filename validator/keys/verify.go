/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

package keys

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tapglue/backend/core"
	. "github.com/tapglue/backend/utils"
	"log"
)

// VerifyRequest verifies if a request is properly signed or not
func VerifyRequest(requestScope, requestVersion string, r *http.Request, numKeyParts int) bool {
	signature := r.Header.Get("x-tapglue-signature")
	if signature == "" {
		log.Printf("signature failed on 1\n")
		return false
	}

	if _, err := Base64Decode(signature); err != nil {
		log.Printf("signature failed on 2\n")
		return false
	}

	payload := PeakBody(r).Bytes()
	if Base64Encode(Sha256String(payload)) != r.Header.Get("x-tapglue-payload-hash") {
		log.Printf("signature failed on 3\n")
		return false
	}

	encodedIds := r.Header.Get("x-tapglue-id")
	decodedIds, err := Base64Decode(encodedIds)
	if err != nil {
		log.Printf("signature failed on 4\n")
		return false
	}

	ids := strings.SplitN(string(decodedIds), ":", numKeyParts)
	if len(ids) != numKeyParts {
		log.Printf("signature failed on 5\n")
		return false
	}

	accountID, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		log.Printf("signature failed on 6\n")
		return false
	}

	authToken := ""
	if numKeyParts == 1 {
		account, err := core.ReadAccount(accountID)
		if err != nil {
			log.Printf("signature failed on 7\n")
			return false
		}
		authToken = account.AuthToken
	} else {
		applicationID, err := strconv.ParseInt(ids[1], 10, 64)
		if err != nil {
			log.Printf("signature failed on 8\n")
			return false
		}

		application, err := core.ReadApplication(accountID, applicationID)
		if err != nil {
			log.Printf("signature failed on 9\n")
			return false
		}

		authToken = application.AuthToken
	}

	signString := generateSigningString(requestScope, requestVersion, r)

	signingKey := generateSigningKey(authToken, requestScope, requestVersion, r)

	/* TODO Debug content, don't remove unless you want to redo it later
	fmt.Printf("\nPayload %s - %s \n", r.Header.Get("x-tapglue-payload-hash"), Base64Encode(Sha256String(payload)))
	fmt.Printf("\nSession %s\n", r.Header.Get("x-tapglue-session"))
	fmt.Printf("\nSignature parts %s - %s \n", Base64Encode(signingKey), Base64Encode(signString))
	fmt.Printf("\nSignature %s - %s \n\n", r.Header.Get("x-tapglue-signature"), Base64Encode(Sha256String([]byte(signingKey+signString))))
	*/

	return r.Header.Get("x-tapglue-signature") == Base64Encode(Sha256String([]byte(signingKey+signString)))
}
