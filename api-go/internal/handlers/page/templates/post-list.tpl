{{define "post-list"}}
{{template "layout" .}}
{{end}}

{{define "head"}}
<style>
  article {
    padding-top: 1rem;
    padding-bottom: 1rem;
  }
  article > a {
    display: block;
    opacity: 1;
    transform: scale(1);
    transition: all .2s ease-in-out;
  }
  article > a:hover {
    transform: scale(1.007);
  }
  h2 {
    font-size: 1.5rem;
    color: hsl(var(--foreground) / 0.93);
  }
  time {
    margin-top: 0.5rem;
    color: hsl(var(--foreground) / 0.80);
    font-size: 0.8rem;
  }
  p {
    margin-top: 0.5rem;
    color: hsl(var(--foreground) / 0.85);
  }
</style>
<title>Pebble</title>
{{end}}

{{define "content"}}
<div class="articles">
  {{range .posts}}
  <article>
    <a href="/shared/{{.ID}}" rel="prefetch">
      <h2>{{if .Title}}{{.Title}}{{else}}Untitled{{end}}</h2>
      <time>{{.CreatedAt}}</time>
      {{if .Description}}
      <p>{{.Description}}</p>
      {{end}}
    </a>
  </article>
  {{end}}
</div>
{{end}}

{{define "scripts"}}{{end}}