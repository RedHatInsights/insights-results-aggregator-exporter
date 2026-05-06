/*
Copyright © 2021, 2022, 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/s3.html

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// error messages
const (
	unableToInitializeConnection = "Unable to initialize connection to S3"
	minioClientIsNil             = "Minio Client is nil"
	wrongMinioClientReference    = "Wrong Minio client reference"
	wrongBucketName              = "Wrong bucket name"
	objectNameIsNotSet           = "Object name is not set"
	wrongObjectName              = "Wrong object name"
	bucketNameIsNotSet           = "Bucket name is not set"
	configurationIsNil           = "Configuration is nil"
	configurationError           = "Configuration error"
)

// NewS3Connection function initializes connection to S3/Minio storage.
func NewS3Connection(configuration *ConfigStruct) (*minio.Client, context.Context, error) {
	// check if configuration structure has been provided
	if configuration == nil {
		err := errors.New(configurationIsNil)
		log.Error().Err(err).Msg(configurationError)
		return nil, nil, err
	}

	// retrieve S3/Minio configuration
	s3Configuration := GetS3Configuration(configuration)

	var endpoint string
	if s3Configuration.EndpointPort == 0 {
		endpoint = s3Configuration.EndpointURL
	} else {
		endpoint = fmt.Sprintf("%s:%d",
			s3Configuration.EndpointURL, s3Configuration.EndpointPort)
	}
	log.Info().Str("S3 endpoint", endpoint).Msg("Preparing connection")

	ctx := context.Background()

	// initialize Minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			s3Configuration.AccessKeyID,
			s3Configuration.SecretAccessKey, ""),
		Secure: s3Configuration.UseSSL,
	})

	// check if client has been constructed properly
	if err != nil {
		log.Error().Err(err).Msg(unableToInitializeConnection)
		return nil, nil, err
	}

	log.Info().Msg("Connection established")
	return minioClient, ctx, nil
}

// s3BucketExists function checks if bucket with given name exists and can be
// accessed by current client
func s3BucketExists(ctx context.Context, minioClient *minio.Client,
	bucketName string) (bool, error) {
	// check if Minio client has been passed to this function
	if minioClient == nil {
		err := errors.New(minioClientIsNil)
		log.Error().Err(err).Msg(wrongMinioClientReference)
		return false, err
	}

	// check if proper bucket name has been passed to this function
	if bucketName == "" {
		err := errors.New(bucketNameIsNotSet)
		log.Error().Err(err).Msg(wrongBucketName)
		return false, err
	}

	// check bucket existence
	found, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Error().Err(err).Str("bucket", bucketName).Msg("Bucket can not be found")
		return false, err
	}

	// everything seems to be ok
	return found, nil
}

// storeTableNames function stores all table names passed via tableNames
// parameter into given bucket under selected object name
func storeTableNames(ctx context.Context, minioClient *minio.Client,
	bucketName string, objectName string, tableNames []TableName) error {
	// check if Minio client has been passed to this function
	if minioClient == nil {
		err := errors.New(minioClientIsNil)
		log.Error().Err(err).Msg(wrongMinioClientReference)
		return err
	}

	// check if proper bucket name has been passed to this function
	if bucketName == "" {
		err := errors.New(bucketNameIsNotSet)
		log.Error().Err(err).Msg(wrongBucketName)
		return err
	}

	// check if proper object name has been passed to this function
	if objectName == "" {
		err := errors.New(objectNameIsNotSet)
		log.Error().Err(err).Msg(wrongObjectName)
		return err
	}

	// conversion to CSV
	buffer := new(bytes.Buffer)

	writer := csv.NewWriter(buffer)
	var data = [][]string{{tableNameMsg}}

	err := writer.WriteAll(data)
	if err != nil {
		return err
	}

	for _, tableName := range tableNames {
		err := writer.Write([]string{string(tableName)})
		if err != nil {
			log.Error().Err(err).Msg("Write to CSV")
		}
	}

	writer.Flush()

	reader := io.Reader(buffer)

	// store CSV data into S3/Minio
	options := minio.PutObjectOptions{ContentType: contentTypeCSV}
	_, err = minioClient.PutObject(ctx, bucketName, objectName, reader, -1, options)
	if err != nil {
		return err
	}

	// everything seems to be ok
	return nil
}

// storeDisabledRulesIntoS3 function stores info about disabled rules into S3
// into given bucket under selected object name
func storeDisabledRulesIntoS3(ctx context.Context, minioClient *minio.Client,
	bucketName string, objectName string, disabledRulesInfo []DisabledRuleInfo) error {
	// check if Minio client has been passed to this function
	if minioClient == nil {
		err := errors.New(minioClientIsNil)
		log.Error().Err(err).Msg(wrongMinioClientReference)
		return err
	}

	// check if proper bucket name has been passed to this function
	if bucketName == "" {
		err := errors.New(bucketNameIsNotSet)
		log.Error().Err(err).Msg(wrongBucketName)
		return err
	}

	// check if proper object name has been passed to this function
	if objectName == "" {
		err := errors.New(objectNameIsNotSet)
		log.Error().Err(err).Msg(wrongObjectName)
		return err
	}

	// conversion to CSV
	buffer := new(bytes.Buffer)
	err := DisabledRulesToCSV(buffer, disabledRulesInfo)
	if err != nil {
		log.Error().Err(err).Msg("Write table name to CSV")
		return err
	}

	reader := io.Reader(buffer)

	// store CSV data into S3/Minio
	options := minio.PutObjectOptions{ContentType: contentTypeCSV}
	_, err = minioClient.PutObject(ctx, bucketName, objectName, reader, -1, options)
	if err != nil {
		return err
	}

	// everything seems to be ok
	return nil
}

func storeBufferToS3(ctx context.Context, minioClient *minio.Client,
	bucketName string, objectName string, buffer bytes.Buffer) error {
	options := minio.PutObjectOptions{ContentType: "text/plain"}
	_, err := minioClient.PutObject(ctx, bucketName, objectName, &buffer, -1, options)
	return err
}
