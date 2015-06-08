#include <QApplication>
#include <QQmlApplicationEngine>
#include <QQmlContext>

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    app.setAttribute(Qt::AA_SynthesizeMouseForUnhandledTouchEvents, false);
    app.setAttribute(Qt::AA_SynthesizeTouchForUnhandledMouseEvents, false);

    QQmlApplicationEngine engine;
    QQmlContext *context = engine.rootContext();

    QString text("Test text");
    context->setContextProperty("myText", text);

    engine.addImportPath("qrc:/");
    engine.load(QUrl(QStringLiteral("qrc:/main.qml")));

    return app.exec();
}
