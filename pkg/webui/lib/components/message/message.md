The `<Message />` component is used as a wrapper around `react-intl`'s
`<FormattedMessage />` component and will make sure that also strings can be
rendered (with a warning). The message can use placeholders which need to be
defined in the `content.values` or `values`-prop. If an element is passed, the
element will be passed through and rendered. No prop is required and the
element will output null if it receives nothing.
