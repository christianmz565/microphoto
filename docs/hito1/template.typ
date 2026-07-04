#let project(
  title: "",
  authors: (),
  course: "",
  group: "",
  teacher: "",
  doc,
) = {
  set text(
    font: "Liberation Serif",
    size: 12pt,
    hyphenate: false,
    lang: "es",
  )

  set page(margin: (x: 1.8cm, y: 2cm))

  set par(
    justify: true,
    first-line-indent: 0pt,
    spacing: 1em,
    leading: 1em,
  )

  set heading(numbering: "1.1.1.")
  show heading: set text(size: 12pt, weight: "bold")

  show figure.caption: it => [
    #strong[#it.supplement #context it.counter.display(it.numbering).] #emph(it.body)
  ]

  align(center)[
    #strong[UNIVERSIDAD NACIONAL DE SAN AGUSTÍN]\
    #strong[FACULTAD DE INGENIERÍA DE PRODUCCIÓN Y SERVICIOS]\
    #strong[ESCUELA PROFESIONAL DE INGENIERÍA DE SISTEMAS]\ \

    #image("img/logo.png", width: 5cm)\

    #strong[PROYECTO FINAL]\
    #title\ \

    #strong[ASIGNATURA]\
    #course\
    #group\ \

    #strong[DOCENTE]\
    MOLINA BARRIGA, MARIBEL\ \

    #strong[INTEGRANTES]\
    #authors.join("\n")\
    \

    #strong[AREQUIPA - PERÚ]\
    #strong[2026]
  ]

  pagebreak()

  set page(
    numbering: "1",
    number-align: center,
  )
  counter(page).update(1)

  set par(leading: 0.9em)

  align(center)[
    #set text(size: 14pt, weight: "bold")
    #block(below: 2em)[#title]
  ]

  doc
}
