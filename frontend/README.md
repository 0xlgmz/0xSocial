# 0xSocial frontend

## Monetization deployment

The frontend reads `/api/monetization/config` and fails closed when the endpoint is unavailable or malformed. A valid enabled response uses this shape:

```json
{
  "enabled": true,
  "clientId": "ca-pub-1234567890123456",
  "placements": {
    "feed": {
      "enabled": true,
      "slotId": "1234567890",
      "format": "fluid",
      "layoutKey": "publisher-generated-layout-key",
      "firstAfterPosts": 4,
      "repeatEveryPosts": 8,
      "maximumAds": 3
    }
  }
}
```

Before enabling this response in production, configure and publish a European regulations message in Google AdSense Privacy & Messaging (or deploy another Google-certified TCF CMP). The CMP must be present on the page and expose the IAB `__tcfapi`; the frontend never creates its own consent value. It waits for the CMP's `CONSENT_API_READY` callback and a valid TCF decision before loading AdSense. The persistent Privacy choices link calls Google's revocation UI.

Use the exact in-feed `slotId`, `format`, and `layoutKey` generated in the publisher dashboard. Confirm the contact addresses and have the policy text reviewed for the production operator and jurisdictions before launch.

## Development

This template provides a minimal setup to get React working in Vite with HMR and some Oxlint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Oxc](https://oxc.rs)
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/)

## React Compiler

The React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

## Expanding the Oxlint configuration

If you are developing a production application, we recommend enabling type-aware lint rules by installing `oxlint-tsgolint` and editing `.oxlintrc.json`:

```json
{
  "$schema": "./node_modules/oxlint/configuration_schema.json",
  "plugins": ["react", "typescript", "oxc"],
  "options": {
    "typeAware": true
  },
  "rules": {
    "react/rules-of-hooks": "error",
    "react/only-export-components": ["warn", { "allowConstantExport": true }]
  }
}
```

See the [Oxlint rules documentation](https://oxc.rs/docs/guide/usage/linter/rules) for the full list of rules and categories.
