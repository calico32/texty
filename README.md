# texty

texty displays text in a window on the desktop. Simple as that.

Heavily inspired by `nwg-wrapper`, it displays GTK windows on the bottom (or any other) layer using gtk-layer-shell, rendering text provided in the configuration file or from the output of a command.

## Configuration

texty's configuration file is located at one of:

- `$XDG_CONFIG_HOME`/texty/texty.kdl
- ~/.config/texty/texty.kdl
- ~/.texty/texty.kdl

Alternatively, you can specify a custom configuration file using the `-c/--config` command line option.

The configuration is written in [KDL](https://kdl.dev), a document language with XML-like semantics.

texty watches the configuration file for changes and automatically restarts itself when it detects a change.

### Example

```kdl
styles "styles.css"

window id=hello {
    text "Hello, world!"
    position top=16 left=16
    style {
        background-color black
        color white
    }
}

window {
    command date +%H:%M:%S
    interval 1 sec
    position bottom=16 center
    style {
        color blue
        font-size 24
    }
}

defaults {
    layer top

    style {
        font-family Inter
        font-size 16
        font-weight bold
    }
}
```

The example above creates two windows: one displaying static text and another showing the current time, updated every second. The `defaults` section applies to all windows unless overridden. A CSS file in specified to apply additional styles globally and using custom IDs.

### Windows

Each `window` must have one and only one of the following:

- `text` - static text to display
- `file` - path to a file containing text to display
- `command` - command to run, output will be displayed

Text from any of these sources can be styled using the `style` property and can also use Pango markup.

When using `file` or `command`, you can specify an `interval` to update the content periodically in the format `[N unit]...`, e.g.
`interval 1 sec` (every second) or `interval 1 hr 30 min` (every 90 minutes).

The `layer` property sets which layer the window will be displayed on. It can be one of `overlay`, `top`, `bottom`, or `background`. The default is `bottom`.

The `position` property sets the position of the window on the screen. It can be specified using the `top`, `bottom`, `left`, and `right` properties, e.g. `top=16 left=16` places the window 16 pixels from the top and left edges of the screen. The special property `center` can be used to center the window on the screen on the unconstrained axis, e.g. `top=16 center` will center the window horizontally while placing it 16 pixels from the top edge.

The `align` property sets the text alignment within the window. It can be one of `left`, `center`, or `right`. The default is `left`, or `center` if the window is centered.

The optional `id` property (specified inline) allows you to assign a custom ID to the window. This can be useful for targeting the window in CSS styles or for other purposes. If not specified, a random ID will be generated.

### Styling

The `style` property allows you to set CSS styles for the window. Each style property is passed verbatim to GTK, so you can use any CSS property supported by GTK and specify the value in any understandable format, **with one exception: properties except for `font-weight` with a single numerical value will be interpreted as pixels**. For example, `font-size 16` will be passed to GTK as `font-size: 16px`. If provided, multiple KDL values passed in the same property, e.g. `font-family Inter Roboto`, will be joined with spaces before they are given to GTK.

If desired, you can alternatively specify styles in string format, as in the example below:

```kdl
window {
    // ...
    style """
        background-color: black;
        color: white;
        font-size: 16px;
        font-family: Inter, Roboto;
    """
}
```

Note that default styles won't be applied to `style` properties specified in this way.

Finally, an arbitrary CSS file can be specified using the top-level `styles` property. This file will be loaded and applied to all windows. Note that by default, each window will have a unique, random ID assigned to it, but specifying a custom `id` in the window's properties will allow you to target it in the CSS file.

### Defaults

The `defaults` section sets default values for all windows. It can contain any properties that can be set on a `window`, such as `layer`, `position`, and `style`. If a property is specified in both the `defaults` section and a specific `window`, the value from the `window` will take precedence.
