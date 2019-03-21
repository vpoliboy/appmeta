/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata



func IsNotFoundError(err error) bool {
	return err == errNotFound
}