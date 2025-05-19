// TODO: merge different methods for one endpoints
local zip(ary) = { [item[0]]: item[1] for item in ary };

{
  // helper schemas
  schemas: {
    string: {
      type: 'string',
    },
  },

  definition(version='3.0.3', info={}, paths=[], schemas=[], responses=[]): {
    openapi: version,
    info: info,
    paths: zip(paths),
    components: {
      schemas: { [schema.name]: schema.schema for schema in schemas },
      responses: zip(responses),
    },
  },

  schema(name, schema): {
    local root = self,

    name: name,
    ref: { '$ref': '#/components/schemas/%s' % (name) },
    contentSchema: {
      'application/json': {
        schema: root.ref,
      },
    },
    schema: schema,
  },

  arrayOfSchema(schema): { type: 'array', items: schema.ref },

  response(statusCode, description, schema): [
    '%s' % (statusCode),
    {
      description: description,
      content: schema.contentSchema,
    },
  ],

  // path is a an array eg ['v1', 'users', '{id}']
  path(method, path, summary, description, schema, tags=[], parameters=[]): [
    '/' + std.join('/', path),
    {
      [method]: {
        summary: summary,
        description: description,
        operationId: '%s_%s' % [method, std.join('_', path)],
        tags: tags,
        parameters: parameters,

        responses: {
          '200': {
            description: 'Successful response',
            content: schema.contentSchema,
          },
        },
      },
    },
  ],

  GET(path, summary, description, schema, tags=[], parameters=[]): $.path('get', path, summary, description, schema, tags, parameters),
}
