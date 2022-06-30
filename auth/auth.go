package auth

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gitlab.com/feedplan-libraries/cache"
	"gitlab.com/feedplan-libraries/constants"
	"gitlab.com/feedplan-libraries/logger"
)

var jwtMiddleWare *jwtmiddleware.JWTMiddleware

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kid string   `json:"kid"`
	X5c []string `json:"x5c"`
}

type CustomClaims struct {
	// Note that the scope can be string or an array
	RawScope json.RawMessage `json:"scope"`
	// Scopes need to be unmarshalled post the initial unmarshalling as we can't be sure of the type
	Scopes []string `json:"-"`
	jwt.StandardClaims
}

func Init() {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			aud := viper.GetString(constants.JwksAudience)
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, errors.New(constants.JwksAudience)
			}
			// Verify 'iss' claim
			iss := viper.GetString(constants.JwksIssuer)
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New(constants.JwksIssuer)
			}

			cert, err := getPemCert(token)
			if err != nil {
				logger.SugarLogger.Warnw("Cannot get pem cert.", "ErrorMessage: ", err.Error())
				return nil, errors.New("cannot get pem cert")
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})

	jwtMiddleWare = jwtMiddleware
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the client secret key
		err := jwtMiddleWare.CheckJWT(c.Writer, c.Request)
		if err != nil {
			// Token not found
			logger.SugarLogger.Errorw("Token not found. Or JWT is invalid.", "CertHeader", c.Request.Header)
			c.Abort()
			c.Writer.WriteHeader(http.StatusUnauthorized)
			_, _ = c.Writer.Write([]byte("Unauthorized"))
			return
		}
	}
}

// IsAuthorizedUser : Checks if the customer id matches the value of subject
// in authorization token.
func IsAuthorizedUser(token, cid string) bool {
	if len(token) == 0 {
		logger.SugarLogger.Warnw("Token cannot be Empty", "CustomerID ", cid)
		return false
	}
	if len(cid) == 0 || cid == uuid.Nil.String() {
		logger.SugarLogger.Warnw("CustomerID cannot be Empty", "CustomerID ", cid)
		return false
	}
	jsonTokens := strings.Split(token, ".")
	if len(jsonTokens) != 3 {
		logger.SugarLogger.Warnw("Unexpected token structure", "Token ", token, "CustomerID ", cid)
		return false
	}

	decodedToken, decodeError := b64.RawStdEncoding.DecodeString(jsonTokens[1])
	if decodeError != nil {
		logger.SugarLogger.Warnw("Unable to decode token", "CustomerID ", cid, "Payload ", jsonTokens[1], "ErrorMessage ", decodeError.Error())
		return false
	}
	claims := CustomClaims{}
	marshallError := json.Unmarshal([]byte(decodedToken), &claims)
	if marshallError != nil {
		logger.SugarLogger.Errorw("Unable to unmarshal decoded claims", "DecodedToken ", string(decodedToken), "CustomerID ", cid)
		return false
	}
	claims.Scopes, marshallError = claims.getScopes()
	if marshallError != nil {
		logger.SugarLogger.Errorw("Unable to get scopes", "CustomerID", cid, "Token", token)
		return false
	}
	for _, s := range claims.Scopes {
		if strings.EqualFold(constants.UserJourneyScope, s) {
			if claims.Subject == "" {
				logger.SugarLogger.Warnw("No subject found in claims even though scope is user_journey", "CustomerID ", cid, "Token ", token)
				return false
			}
			// returning if subject matches customer id
			return strings.EqualFold(claims.Subject, cid)
		}
	}
	logger.SugarLogger.Debugw("Cid claim is valid as not a customer token", "CustomerID ", cid)
	// Returning true if in-case it is not a customer token
	return false
}

func ValidateScope(token, validScope string) bool {
	jsonTokens := strings.Split(token, ".")
	if len(jsonTokens) != 3 {
		logger.SugarLogger.Warnw("Token structure does not seem to be as expected. Token: scope", "token", token)
		return false
	}

	payloadToken := jsonTokens[1]
	decodedToken, decodeError := b64.StdEncoding.DecodeString(payloadToken + "==")
	if decodeError != nil {
		logger.SugarLogger.Warnw("Unable to decode token. Payload token:", "TokenPayload", payloadToken, "ErrorMessage", decodeError.Error())
		return false
	}

	claims := CustomClaims{}
	marshallError := json.Unmarshal([]byte(decodedToken), &claims)
	if marshallError != nil {
		logger.SugarLogger.Errorw("Unable to unmarshal decoded claims.", "DecodedToken:", decodedToken)
		return false
	}

	claims.Scopes, marshallError = claims.getScopes()
	if marshallError != nil {
		logger.SugarLogger.Errorw("Unable to get scopes. for token ", "error ", marshallError)
		return false
	}

	for _, scope := range claims.Scopes {
		if strings.Contains(validScope, scope) {
			return true
		}
	}
	// Returning true if in-case it is not a customer token
	return false
}

// ValidateCustomerTokenWithID : Checks if the customer id matches the value of subject in authorization token.
func ValidateCustomerTokenWithID(token string, cid string) bool {
	if len(token) == 0 {
		logger.SugarLogger.Warnw("Token cannot be Empty", "CustomerID", cid)
		return false
	}

	if len(cid) == 0 || cid == uuid.Nil.String() {
		logger.SugarLogger.Warnw("CustomerID cannot be Empty", "CustomerID", cid)
		return false
	}

	jsonTokens := strings.Split(token, ".")
	if len(jsonTokens) != 3 {
		logger.SugarLogger.Warnw("Unexpected token structure", "Token", token, "CustomerID", cid)
		return false
	}

	decodedToken, decodeError := b64.StdEncoding.DecodeString(jsonTokens[1] + "==")
	if decodeError != nil {
		logger.SugarLogger.Warnw("Unable to decode token", "CustomerID", cid, "Payload", jsonTokens[1], "ErrorMessage", decodeError.Error())
		//return false
	}

	claims := CustomClaims{}
	marshallError := json.Unmarshal(decodedToken, &claims)
	if marshallError != nil {
		logger.SugarLogger.Warnw("Unable to unmarshal decoded claims", "DecodedToken", string(decodedToken), "CustomerID", cid)
		return false
	}

	claims.Scopes, marshallError = claims.getScopes()
	if marshallError != nil {
		logger.SugarLogger.Warnw("Unable to get scopes", "CustomerID", cid, "Token", token)
		return false
	}

	for _, s := range claims.Scopes {
		if strings.EqualFold("user_journey", s) {
			if claims.Subject == "" {
				logger.SugarLogger.Warnw("No subject found in claims even though scope is user_journey", "CustomerID", cid, "Token", token)
				return false
			}
			// returning if subject matches customer id
			return strings.EqualFold(claims.Subject, cid)
		}
	}

	logger.SugarLogger.Debugw("Cid claim is valid as not a customer token", "CustomerID", cid)
	// Returning true if in-case it is not a customer token
	return true
}

func (claims *CustomClaims) getScopes() ([]string, error) {
	if len(claims.RawScope) == 0 {
		logger.SugarLogger.Warnw("Scope raw message is empty.", "Claims ", claims)
		return nil, errors.New("scope raw message is empty")
	}

	switch claims.RawScope[0] {
	case '"':
		var scope string
		if err := json.Unmarshal(claims.RawScope, &scope); err != nil {
			logger.SugarLogger.Errorw("Unable to unmarshal stringified scope.", "RawScope: ", claims.RawScope)
			return nil, errors.New("unable to unmarshall string scope")
		}
		return []string{scope}, nil

	case '[':
		var scopes []string
		if err := json.Unmarshal(claims.RawScope, &scopes); err != nil {
			logger.SugarLogger.Errorw("Unable to unmarshal arrayed scopes.", "RawScopes: ", claims.RawScope)
			return nil, errors.New("unable to unmarshall arrayed scopes")
		}
		return scopes, nil
	}
	logger.SugarLogger.Warnw("Unable to unmarshal scopes.", "RawScopes: ", claims.RawScope)
	return nil, errors.New("unable to unmarshal scopes")
}

func getPemCert(token *jwt.Token) (string, error) {
	redisClient := cache.GetRedisClientImp()
	environment := viper.GetString(constants.Environment)
	redisKey := constants.ServiceNameKey + constants.ColonSeparatorForRedisKey + environment + constants.ColonSeparatorForRedisKey + constants.JwksResponseKey
	jwksResponse := Jwks{}
	cert := ""

	cachedResponse, cachedResponseErr := redisClient.Get(redisKey)

	if cachedResponseErr == nil && len(cachedResponse) > 0 {
		logger.SugarLogger.Errorw("Found JWKS response in redis cache. Un-marshaling the response", "CachedResponse", cachedResponse)
		cachedResponseErr = json.Unmarshal([]byte(cachedResponse), &jwksResponse)
		if cachedResponseErr != nil {
			logger.SugarLogger.Warnw("Failed to unmarshal JWKS response from redis cache. Hence, calling JwksUrl to get the value", "CachedResponse", cachedResponse, "ErrorMessage", cachedResponseErr.Error())
		}
	}
	if cachedResponseErr != nil || len(cachedResponse) == 0 {
		jwksResponse = Jwks{}
		resp, err := http.Get(viper.GetString(constants.JwksUrl))
		if err != nil {
			return cert, err
		}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&jwksResponse)
		if err != nil {
			return cert, err
		}
		jwksResponseInBytes, marshalErr := json.Marshal(jwksResponse)
		if marshalErr == nil {
			_, err = redisClient.Set(redisKey, jwksResponseInBytes, constants.JwksResponseCacheTimeout)
			if err != nil {
				logger.SugarLogger.Errorw("Failed to cache JWKS response in redis", "ErrorMessage", err.Error())
			}
		} else {
			logger.SugarLogger.Errorw("Failed to marshal JWKS response. Hence, could not set value in redis cache", "ErrorMessage", marshalErr)
		}
	}

	for k := range jwksResponse.Keys {
		if token.Header[constants.Kid] == jwksResponse.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwksResponse.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		logger.SugarLogger.Infow("Kid from token header : ", token.Header["kid"])
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

func RemoveJwksCache() error {
	redisClient := cache.GetRedisClientImp()
	environment := viper.GetString(constants.Environment)
	jwksCacheRedisKey := constants.ServiceNameKey + constants.ColonSeparatorForRedisKey + environment + constants.ColonSeparatorForRedisKey + constants.JwksResponseKey
	_, err := redisClient.Del(jwksCacheRedisKey)
	if err != nil {
		logger.SugarLogger.Errorw("Unable to remove jwks cache", "ErrorMessage", err.Error(), "Key", jwksCacheRedisKey)
	}
	return err
}
