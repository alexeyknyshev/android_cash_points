TEMPLATE = app

QT += qml quick location sql svg

android: QT += androidextras

SOURCES += src/main.cpp \
    src/banklistsqlmodel.cpp \
    src/townlistsqlmodel.cpp \
    src/serverapi.cpp \
    src/listsqlmodel.cpp \
    src/icoimageprovider.cpp \
    src/searchengine.cpp \
    src/locationservice.cpp \
    src/cashpointsqlmodel.cpp \
    src/requests/cashpointrequest.cpp \
    src/requests/cashpointinradius.cpp \
    src/requests/requestfactory.cpp \
    src/requests/cashpointrequestinradiusfactory.cpp \
    src/cashpointresponse.cpp \
    src/requests/cashpointcreate.cpp \
    src/requests/nearbyclusters.cpp \
    src/requests/nearbyclusterrequestfactory.cpp \
    src/emptyimageprovider.cpp \
    src/appstateproxy.cpp \
    src/hostsmodel.cpp \
    src/requests/cashpointedit.cpp \
    src/requests/cashpointpatches.cpp \
    src/requests/cashpointcreatefactory.cpp \
    src/requests/cashpointeditfactory.cpp \
    src/requests/cashpointpatchesfactory.cpp \
    src/feedbackservice.cpp \
    src/googleapiclient.cpp \
    src/filtersmodel.cpp

RESOURCES += qml.qrc

# Additional import path used to resolve QML modules in Qt Creator's code model
QML_IMPORT_PATH =

# Default rules for deployment.
include(deployment.pri)

DISTFILES += \
    android/AndroidManifest.xml \
    android/gradle/wrapper/gradle-wrapper.jar \
    android/gradlew \
    android/res/values/libs.xml \
    android/build.gradle \
    android/gradle/wrapper/gradle-wrapper.properties \
    android/gradlew.bat

ANDROID_PACKAGE_SOURCE_DIR = $$PWD/android

HEADERS += \
    src/banklistsqlmodel.h \
    src/townlistsqlmodel.h \
    src/serverapi.h \
    src/rpctype.h \
    src/listsqlmodel.h \
    src/icoimageprovider.h \
    src/searchengine.h \
    src/locationservice.h \
    src/cashpointsqlmodel.h \
    src/requests/cashpointrequest.h \
    src/requests/cashpointinradius.h \
    src/requests/requestfactory.h \
    src/requests/cashpointrequestinradiusfactory.h \
    src/serverapi_fwd.h \
    src/cashpointresponse.h \
    src/requests/cashpointcreate.h \
    src/requests/nearbyclusters.h \
    src/requests/nearbyclusterrequestfactory.h \
    src/appstateproxy.h \
    src/hostsmodel.h \
    src/requests/cashpointedit.h \
    src/requests/cashpointpatches.h \
    src/requests/cashpointcreatefactory.h \
    src/requests/cashpointeditfactory.h \
    src/requests/cashpointpatchesfactory.h \
    src/feedbackservice.h \
    src/googleapiclient.h \
    src/filtersmodel.h

OTHER_FILES += \
    android/src/net/agnia/cashpoints/CashPointsActivity.java

QMAKE_CXXFLAGS += -std=c++11

debug {
    DEFINES += CP_DEBUG
}
