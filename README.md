# Disgo

Disgo is a simple commenting system for the web, written in Go. It is inspired by [Disqus](http://disqus.com) but does not come with all the bells and whistles. Some (anti-)features are:

- Ajax-based: no Iframes
- no Javascript dependency, especially not jQuery
- works on all [modern browsers](http://caniuse.com/cors) and Internet Explorer >=10
- supports MySQL and PostgreSQL
- customizable CSS (in fact, there is currently no default client CSS)

Optional:

- admin approval of new comments
- e-mail notification of new comment
- IP address-based rate limiting
- markdown support
- Disqus import

Here's what the admin interface looks like:
![Disgo Admin Interface](http://pascalj.github.io/disgo.png)


## Installation

### Binary release

You can [download a ready-to-go release](https://github.com/pascalj/disgo/releases) for most common platforms (FreeBSD, Linux, OS X, Windows). Just unpack and [configure](#configuration) it.

### Manual build

You need to have the Go environment installed. To build and install disgo, simply run:

```
$ go get github.com/pascalj/disgo
$ go install github.com/pascalj/disgo
```

You can build a release with

```
$ make release
```

`build/` then contains all you need: binaries for the most common platforms, the public files, templates and a sample of the configuration file.

## Configuration

The various switches in the [configuration sample file](disgo.gcfg.sample) are documented. To get started, just copy the `disgo.gcfg.sample` to `disgo.gcfg` (yes, it's `.gcfg` and not `.cfg`) and edit it to your needs.

Two things are mandatory:

A database must be configured, either MySQL or PostgreSQL and you need to set at least one `origin`. An origin is a URL where you want to allow comments. It may contain `*` as a wildcard. You can tell disgo to load a specific config by providing the `-config` command line parameter. It is strongly suggested to change the `secret`.

The server will listen on `0.0.0.0:3000`. You can change that by setting the `HOST` and/or `PORT` environment variable.

Fire up the server process:

```
$ disgo
```

After that, the admin interface will be available at the host and port you configured (e.g. http://example.com:3000). The first time you access the admin panel you're asked to create an admin user.

## Embedding comments

Once you got everything up and running, it's easy to embedd comments in your website. Just add the following script to your website, just before the closing `</body>` tag:

```html
<script>
var disgo = {
    base: 'http://example.com:3000'
}
document.write('<script src="' + disgo.base + '/js/disgo.js">\x3C/script>')
</script>
```

Replace `example.com:3000` with the location of your Disgo installation (or copy it from the admin interface, where it should be correctly displayed).

To display a form and comments, add a `div` with a `data-disgo-url` attribute:

```html
<div data-disgo-url="http://example.com/2014/04/01/facebook-buys-golang"></div>
```

The `data-disgo-url` does not need to be the current URL, it does not even need to be a URL. However, it is used to identify the content that the comments belong to. So if you're hosting a blog and have comments on your posts, it's recommended to use the URL of the posts as `data-disgo-url`.

## Client callbacks

The Javascript Disgo provides works but you might want to tweak it a little bit. For example, you might want to display fancy error messages or animate a new comment. For that purpose there are some callbacks:

- `onSubmitError(status, result, xhr, form)`
- `onSubmitSuccess(status, result, xhr, form)`

Both replace the default actions that do not do anything fancy whatsoever. `onSubmitError` gets called when the disgo server rejects a new comment, for example because the validation did not complete successfully. `onSubmitSuccess` gets called once the comment got saved. `result` will contain the server's answer.

To configure the callbacks, simply set them in the `disgo` object that also configures the base URL:

```javascript
var disgo = {
    base: 'http://example.com:3000',
    onSubmitSuccess: function(status, result, xhr, form) {
		console.log(result)
    }
}
...
```

## Templates

If you're not happy with the pretty generic HTML templates, you can specify a different `template` directory in the config file. Just copy the default templates to a new location, edit them to your needs and adjust the `template` path.

## Disqus import

Disgo can import your existing comments from Disqus. Just [export](http://disqus.com/admin/discussions/export/) all comments and execute:

```
$ disgo -import path/to/disqus.xml
```

## Known issues

*Only one admin is configurable.* There is no user management interface, yet.

## Contributing

Any feedback is welcome. If you have features/suggestions, please add them to the [Trello board](https://trello.com/b/HU7Vc3NT/disgo) or drop me a mail. If you find a bug, please open an issue on Github. Of course I'm happy to merge pull requests.
