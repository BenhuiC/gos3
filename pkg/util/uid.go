package util

import "github.com/google/uuid"

// 1c38d928-c483-4f61-86fb-282e2213f85f
func UUID32() (id string) {
	id = uuid.NewString()
	id = id[0:8] + id[9:13] + id[14:18] + id[19:23] + id[24:]
	return
}
