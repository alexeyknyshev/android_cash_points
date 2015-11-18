#ifndef SERVERAPI_FWD_H
#define SERVERAPI_FWD_H

#include <QtCore/QMetaType>

class ServerApi;

struct ServerApiPtr
{
    ServerApiPtr();
    ServerApiPtr(ServerApi *api);
    ServerApiPtr(const ServerApiPtr &ptr);

    ServerApi* operator ->();
    ServerApi* operator *() const;

    ServerApi *d;
};

Q_DECLARE_METATYPE(ServerApiPtr)

#endif // SERVERAPI_FWD_H

