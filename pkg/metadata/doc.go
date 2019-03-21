/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

// metadata package implements the application metadata indexing service.
// The service provides the following capabilities to the user
//    1. Users can upload application metadata for indeing
//    2. Users can search for metadata that matches the specified filters
//
// The service is created by calling  metadata.NewService(logger *logrus.Logger, opts ...ServiceOption)
// Currently, just one ServiceOption is supported which is the ability to specify custom tokenizers for the search fields.

// For more details on tokenizers, please take a look at analysis.go and tokenizer.go
package metadata
