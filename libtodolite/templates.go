package libtodolite

/*
type TodoliteChanges struct {
	Changes      []TodoliteChange `json:"results"`
	LastSequence interface{}      `json:"last_seq"`
}

type TodoliteChange struct {
	Sequence    interface{}        `json:"seq"`
	Id          string             `json:"id"`
	ChangedRevs []couch.ChangedRev `json:"changes"`
	Deleted     bool               `json:"deleted"`
	Type        string
	Title       string
	Parent      string // The parent list, or N/A
}

*/

const changesTemplate = `<h2>Changes up to sequence: {{.LastSequence}}</h2>
<table border="1">
<tr>
<th>Seq</th>
<th>Type</th>
<th>Id</th>
<th>Deleted</th>
<th>Title</th>
<th>Parent</th>
<th>ChangedRevs</th>
</tr>
{{with .Changes}}
    {{range .}}
        <tr>
        <td>{{.Sequence}}</td>
        <td>{{.Type}}</td>
        <td>{{.Id}}</td>
        <td>{{.Deleted}}</td>
        <td>{{.Title | Truncate }}</td>
        <td>{{.Parent | Truncate }}</td>
        <td>{{.ChangedRevs }}</td>
        </tr>
    {{end}}
{{end}}
</table>

`
