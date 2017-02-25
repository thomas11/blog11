# Blog11

A static site/blog generator written in Go. I wrote this several years back mainly as a project to
learn Go. The code is a mess, mainly the fact that the templates live in another repository but are
tightly coupled to this package.

You're very welcome to use this for your site, but I'd recommend looking at a more mature and
configurable project such as [Hugo](https://gohugo.io). If you're still here, have a look at my
repository thomaskappler.net for a sample configuration using blog11.

# TODO

- serve.go should be part of the package
- myblog.go:main should be part of the package
- When Go 1.8 is released, replace the annoying sorting interfaces
- Needs a command to create a new empty post based on SiteConf