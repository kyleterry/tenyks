def admin_middlware(connection, data):
    conf = connection.config
    data['admin'] = data['nick'] in conf['admins']
    return data


CORE_MIDDLEWARE = (
    admin_middlware,
)
