> [!NOTE]
> This is mirrored from my private forgejo instance

# Zitadel Glauth Plugin

This is a very simple, and very YOLO plugin for Zitadel.

## How does it work?

Querying the Zitadel API for every single LDAP query is pretty slow.

So it basically shlurps users and their grants and metadata into memory and then operates on that.
The in memory data is updated either:

* every 10 minutes
* everytime a bind happens (async)

Also because a password check is pretty slow via the session API and every single LDAP query will do a bind, the password and username pairs are hashed and cached for 1h.

## How to use it?

Create a service account, give it `Org User Manager`. Then create a personal access token for that account.

Set these env variables:

```
export ZITADEL_URL=https://...
export ZITADEL_PAT=....
```

The plugin can handle capabilities and custom fields.

To add capability `search *` to a user, set metadata `cap_search` with the value of `*`.

To add a custom field called `test` to a user, set the metadata `gl_test` with any desired value.

The plugin will only show human users in the LDAP queries!

> [!CAUTION]
> Make sure that you don't have duplicated role names in your projects! The plugin will put them all in the same ou, so they can't have the same names!

## Disclaimer

Use this thing on your own risk! I'm not responsible if it eats your cat or let's all the hackers into your systems.
