// A private location: the demo's own network, where a container of the
// owner's runs the checks in place of the public locations. It exists only when
// CHECKLY_PRIVATE_LOCATION_SLUG names its slug, and the groups read this
// export: with a location here each group runs from it and from no public
// location, without one each runs from its own locations as before. This is
// the Team-plan route, untested on the free tier.

import { PrivateLocation } from 'checkly/constructs'

export let privateLocation: PrivateLocation | undefined

const slug = process.env.CHECKLY_PRIVATE_LOCATION_SLUG
if (slug !== undefined && slug !== '') {
  privateLocation = new PrivateLocation('demo-private', {
    name: 'The demo network',
    slugName: slug,
    icon: 'server',
  })
}
