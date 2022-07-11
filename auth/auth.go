package auth

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/form3tech-oss/jwt-go"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"gitlab.com/feedplan-libraries/common/cache"
	"gitlab.com/feedplan-libraries/common/constants"
	"gitlab.com/feedplan-libraries/common/logger"
)

var (
	jwtMiddleWare *jwtmiddleware.JWTMiddleware

	jwksUrl = "https://feedplan-" + viper.GetString(constants.Environment) + ".s3." + viper.GetString(constants.AwsRegionKey) + ".amazonaws.com/jwks"
)

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type CustomClaims struct {
	// Note that the scope can be string or an array
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

	// returning if subject matches customer id
	return strings.EqualFold(claims.Subject, cid)

}

func getPemCert(token *jwt.Token) (string, error) {
	redisClient := cache.GetRedisClientImp()
	environment := viper.GetString(constants.Environment)
	serviceName := viper.GetString(constants.ServiceNameKey)
	redisKey := serviceName + constants.ColonSeparatorForRedisKey + environment + constants.ColonSeparatorForRedisKey + constants.JwksResponseKey
	var jwksResponse = Jwks{}
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

		resp, err := http.Get(viper.GetString(jwksUrl))
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
