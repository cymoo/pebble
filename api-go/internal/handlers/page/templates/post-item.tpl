{{define "post-item"}}
{{template "layout" .}}
{{end}}

{{define "head"}}
<link href="/static/photoswipe.css" rel="stylesheet"/>
<style>
  .gallery {
    margin-top: 1rem;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 0.5rem;
  }
  @media (min-width: 640px) {
    .gallery {
      grid-template-columns: repeat(5, minmax(0, 1fr));
    }
  }
  .gallery a {
    position: relative;
    aspect-ratio: 1 / 1;
    transition: transform 0.3s ease;
  }
  .gallery a:hover {
    transform: scale(1.05);
  }
  .gallery a:focus img {
    border: 1px solid hsl(174 42% 65%);
  }
  .gallery img {
    display: inline-block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
</style>
<title>{{.title}}</title>
{{end}}

{{define "content"}}
<article class="prose">
  {{.post.Content}}
</article>
{{if .images}}
  <div class="gallery">
    {{range .images}}
      <a data-cropped="true"
         {{if .Height}}data-pswp-height="{{.Height}}"{{end}}
         {{if .Width}}data-pswp-width="{{.Width}}"{{end}}
         href="{{.URL}}"
         rel="noreferrer"
         tabindex="0"
         target="_blank"
      >
        <img alt="" src="{{if .ThumbURL}}{{.ThumbURL}}{{else}}{{.URL}}{{end}}">
      </a>
    {{end}}
  </div>
{{end}}
{{end}}

{{define "scripts"}}
<script type="module">
  import PhotoSwipeLightbox from "/static/photoswipe-lightbox.esm.min.js"
  const lightbox = new PhotoSwipeLightbox({
    gallery: '.gallery',
    children: 'a',
    pswpModule: () => import("/static/photoswipe.esm.min.js")
  });
  lightbox.init();
</script>
{{end}}