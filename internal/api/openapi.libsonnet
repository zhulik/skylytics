// TODO: merge different methods for one endpoints
local zip(ary) = { [item[0]]: item[1] for item in ary };

local openapi = {
  definition(version='3.0.3', info={}, paths=[], schemas=[], responses=[]): {
    openapi: version,
    info: info,
    paths: zip(paths),
    components: {
      schemas: zip(schemas),
      responses: zip(responses),
    },
  },

  schemaRef(name): { '$ref': '#/components/schemas/%s' % (name) },

  contentSchema(name): {
    'application/json': {
      schema: openapi.schemaRef(name),
    },
  },

  // path is a an array eg ['v1', 'users', '{id}']
  path(method, path, summary, description, schemaName): [
    '/' + std.join('/', path),
    {
      [method]: {
        summary: summary,
        description: description,
        operationId: '%s_%s' % [method, std.join('_', path)],
        responses: {
          '200': {
            description: 'Successful response',
            content: openapi.contentSchema('Posts'),
          },
        },
      },
    },
  ],

  schema(name, schema): [
    name,
    schema,
  ],

  arrayOfSchema(itemSchemaName): {
    type: 'array',
    items: openapi.schemaRef(itemSchemaName),
  },

  response(statusCode, description, schemaName): [
    '%s' % (statusCode),
    {
      description: description,
      content: openapi.contentSchema(schemaName),
    },
  ],
};

openapi
