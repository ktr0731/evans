// Package cache provides read/write access to the Evans's cache file.
//
// All APIs which modify each config elements are provided as top-level functions
// which begin with Set*. Note that all changes by Set* functions aren't flushed
// until calling Save() of the returned cache object.
package cache
