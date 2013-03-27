def admin_middlware(connection, data):
    conf = connection.connection_config
    data['admin'] = data['nick'] in conf['admins']
    return data
