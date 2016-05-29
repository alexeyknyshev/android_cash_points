#include "appstateproxy.h"

#include <QtGui/QGuiApplication>
#include <QtCore/QDebug>

AppStateProxy::AppStateProxy(QGuiApplication *app)
    : QObject(app)
{
    connect(app, SIGNAL(applicationStateChanged(Qt::ApplicationState)),
            this, SLOT(onAppStateChanged(Qt::ApplicationState)));
}

void AppStateProxy::onBanksDataLoaded()
{
    qDebug() << "banks data loaded";
    emit serverDataLoaded(true, "banks");
}

void AppStateProxy::onTownsDataLoaded()
{
    qDebug() << "towns data loaded";
    emit serverDataLoaded(true, "towns");
}

void AppStateProxy::onConnectionFailed()
{
    emit serverDataLoaded(false, "");
}

void AppStateProxy::onAppStateChanged(Qt::ApplicationState state)
{
    emit appStateChanged((int)state);
}

