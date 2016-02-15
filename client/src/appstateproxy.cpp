#include "appstateproxy.h"

#include <QtGui/QGuiApplication>

AppStateProxy::AppStateProxy(QGuiApplication *app)
    : QObject(app)
{
    connect(app, SIGNAL(applicationStateChanged(Qt::ApplicationState)),
            this, SLOT(onAppStateChanged(Qt::ApplicationState)));
}

void AppStateProxy::onAppStateChanged(Qt::ApplicationState state)
{
    emit appStateChanged((int)state);
}

