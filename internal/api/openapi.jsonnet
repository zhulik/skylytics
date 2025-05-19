local openapi = import 'openapi.libsonnet';

openapi.definition(
  paths=[
    openapi.path(
      'get',
      ['v1', 'posts'],
      summary='List posts',
      description='Retrieve posts',
      schemaName='Posts',
    ),
  ],

  schemas=[
    openapi.schema('Error', (import 'schemas/error.libsonnet')),

    openapi.schema('DID', (import 'schemas/did.libsonnet')),
    openapi.schema('DIDs', openapi.arrayOfSchema('DID')),

    openapi.schema('Post', (import 'schemas/post.libsonnet')),
    openapi.schema('Posts', openapi.arrayOfSchema('Post')),
  ],

  responses=[
    openapi.response(400, 'Bad request', 'Error'),
    openapi.response(404, 'Not found', 'Error'),
    openapi.response(500, 'Internal server error', 'Error'),
  ]
)
