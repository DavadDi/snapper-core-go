package util

// IDToSource ...
func IDToSource(ID string) string {
	switch ID[0:2] {
	case "a.":
		return "android"
	case "i.":
		return "ios"
	}
	return "web"
}

// func  SessionToID (session) {
//   let source = (session.source || 't').toLowerCase()
//   source = source[0]
//   if (!'aiwt'.includes(source)) source = 't'

//   let id = `${source}.`
//   let timestamp = session.exp.toString(16)
//   if (timestamp.length % 2) timestamp = '0' + timestamp
//   id += new Buffer(session.userId + timestamp, 'hex').toString('base64')
//   return id.replace(/\//g, '_').replace(/\+/g, '-').replace(/=/g, '~')
// }
