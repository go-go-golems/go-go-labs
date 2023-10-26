package web

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-go-golems/go-go-labs/cmd/tools/create-book-db/chapters"
	"html/template"
	"net/http"
	"strconv"
)

const indexTemplate = `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="//fonts.googleapis.com/css?family=Roboto:300,300italic,700,700italic">
    <link rel="stylesheet" href="//cdn.rawgit.com/necolas/normalize.css/master/normalize.css">
    <link rel="stylesheet" href="//cdn.rawgit.com/milligram/milligram/master/dist/milligram.min.css">
</head>
<body>
    <table>
        <tr>
            <th>Title</th>
            <th>Filename</th>
        </tr>
        {{range .}}
            <tr>
                <td><a href="/chapters/{{.ID}}">{{.Title}}</a></td>
                <td>{{.Filename}}</td>
            </tr>
        {{end}}
    </table>
</body>
</html>
`

const chapterTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="//fonts.googleapis.com/css?family=Roboto:300,300italic,700,700italic">
    <link rel="stylesheet" href="//cdn.rawgit.com/necolas/normalize.css/master/normalize.css">
    <link rel="stylesheet" href="//cdn.rawgit.com/milligram/milligram/master/dist/milligram.min.css">
</head>
<body>
    <h1>{{.Title}}</h1>
    <p><strong>ID:</strong> {{.ID}}</p>
    <p><strong>Filename:</strong> {{.Filename}}</p>
    <p><strong>Created At:</strong> {{.CreatedAt}}</p>
    <p><strong>Comment:</strong> {{.Comment}}</p>
    <h2>Summary</h2>
    <p>{{.SummaryParagraph}}</p>

    <h2>Table of Contents</h2>
    <ul>
        {{range .Toc}}
        <li>{{.Name}} (Level: {{.Level}})</li>
        {{end}}
    </ul>

    <h2>Key Points</h2>
    <ul>
        {{range .KeyPoints}}
        <li>{{.}}</li>
        {{end}}
    </ul>

    <h2>Keywords</h2>
    <ul>
        {{range .Keywords}}
        <li>{{.}}</li>
        {{end}}
    </ul>

    <h2>Sources</h2>
    <ul>
        {{range .Sources}}
        <li>Author: {{.Author}}, Source: {{.Source}}</li>
        {{end}}
    </ul>
</body>
</html>
`

func Serve(db *sql.DB) error {
	r := gin.Default()
	r.GET("/chapters/:id", func(c *gin.Context) {
		// Extract chapter ID from URL
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter ID"})
			return
		}

		// Get the chapter by ID
		chapter, err := chapters.GetChapterByID(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
			return
		}

		// Load HTML template
		tmpl := template.Must(template.New("chapter").Parse(chapterTemplate))

		// Render template
		err = tmpl.Execute(c.Writer, chapter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})
	r.GET("/", func(c *gin.Context) {
		chapters, err := chapters.GetAllChapters(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		tmpl := template.Must(template.New("chapters").Parse(indexTemplate))
		err = tmpl.Execute(c.Writer, chapters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})

	return r.Run(":8080")
}
