// Package all は全エンジンを blank import して init() 登録を集約する。
// bench / cmd はこれを import することで engine.New が全エンジンを引ける。
// 新エンジン追加時はここに1行足す。
package all

import (
	_ "polystore_database/src/go/engine/bulk"
	_ "polystore_database/src/go/engine/stream"
	_ "polystore_database/src/go/engine/vecstream"
	_ "polystore_database/src/go/engine/volcano"
)
