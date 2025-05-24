local oapi = import 'openapi.libsonnet';

local schemas = {
  Error: oapi.schema('Error', (import 'schemas/error.libsonnet')),

  DID: oapi.schema('DID', (import 'schemas/did.libsonnet')),
  DIDs: oapi.schema('DIDs', oapi.arrayOfSchema(self.DID)),

  Post: oapi.schema('Post', (import 'schemas/post.libsonnet')),
  Posts: oapi.schema('Posts', oapi.arrayOfSchema(self.Post)),

  Object: oapi.schema('Object', { type: 'object', additionalProperties: true }),
};

oapi.definition(
  paths=[
    oapi.GET(['v1', 'openapi'], 'This openapi definition', 'This openapi definition', schemas.Object, tags=['openapi']),
    oapi.GET(['v1', 'posts'], 'List posts', 'Retrieve posts', schemas.Posts, tags=['posts']),
    oapi.GET(['v1', 'posts', '{id}'], 'Get post', 'Retrieve post', schemas.Post, tags=['posts'], parameters=[
      { name: 'id', 'in': 'path', required: true, schema: oapi.schemas.string },
    ]),
  ],

  schemas=std.objectValues(schemas),

  responses=[
    oapi.response(400, 'Bad request', schemas.Error),
    oapi.response(404, 'Not found', schemas.Error),
    oapi.response(500, 'Internal server error', schemas.Error),
  ]
)
