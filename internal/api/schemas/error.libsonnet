{
  type: 'object',
  properties: {
    code: {
      description: 'Error code',
      type: 'string',
      example: [
        'not_found',
        'internal_server_error',
        'validation_failed',
      ],
    },
    message: {
      description: 'Error message',
      type: 'string',
      example: [
        'Requested user not found',
        'Internal server error',
        'Validation failed',
      ],
    },
    message_localized: {
      description: 'Human readable error message.',
      type: 'string',
      example: [
        'Requested user not found',
        'Запрашиваемый пользователь не найден.',
      ],
    },
  },
  required: [
    'code',
    'message',
    'message_localized',
  ],
}
