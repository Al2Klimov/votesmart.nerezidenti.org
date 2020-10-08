import React, {Component, ReactNode} from 'react';

import {
  AsyncStorage,
  Button,
  SafeAreaView,
  StyleSheet,
  ScrollView,
  View,
  Text,
  StatusBar,
} from 'react-native';

import {Colors} from 'react-native/Libraries/NewAppScreen';

type State = {view: 'is-russian' | 'not-russian'};

export default class App extends Component<{}, State> {
  persistState(state: State) {
    AsyncStorage.setItem('state', JSON.stringify(state)).then(
      () => {
        this.setState(state);
      },
      (reason) => {
        throw reason;
      },
    );
  }

  initState() {
    this.persistState({view: 'is-russian'});
  }

  render(): ReactNode {
    const sections: ReactNode[] = [];

    if (this.state === null) {
      AsyncStorage.getItem('state').then(
        (state) => {
          if (state === null) {
            this.initState();
          } else {
            let loaded: any;

            try {
              loaded = JSON.parse(state);
            } catch (e) {
              this.initState();
              return;
            }

            this.setState(loaded);
          }
        },
        (reason) => {
          throw reason;
        },
      );
    } else {
      switch (this.state.view) {
        case 'is-russian':
          sections.push(
            <>
              <View style={styles.sectionContainer}>
                <Text style={styles.sectionTitle}>Вы –</Text>
                <Text style={styles.sectionDescription}>
                  гражданин Российской Федерации?
                </Text>
                <Text style={styles.sectionDescription} />
                <Button key="is-russian-yes" title="Да" onPress={() => {}} />
                <Text style={styles.sectionDescription} />
                <Button
                  title="Нет"
                  key="is-russian-no"
                  onPress={() => {
                    this.persistState({view: 'not-russian'});
                  }}
                />
              </View>
            </>,
          );
          break;
        case 'not-russian':
          sections.push(
            <>
              <View style={styles.sectionContainer}>
                <Text style={styles.sectionTitle}>Вы –</Text>
                <Text style={styles.sectionDescription}>
                  не гражданин Российской Федерации.
                </Text>
                <Text style={styles.sectionDescription} />
                <Button
                  key="not-russian-back"
                  title="Назад"
                  onPress={() => {
                    this.persistState({view: 'is-russian'});
                  }}
                />
                <Text style={styles.sectionDescription} />
                <Text style={styles.sectionDescription}>
                  К сожалению, Вам придётся бороться с ОПГ «Единая Россия»
                  как-то по-другому – это приложение Вам в этом не поможет.
                </Text>
              </View>
            </>,
          );
      }
    }

    return (
      <>
        <StatusBar barStyle="dark-content" />
        <SafeAreaView>
          <ScrollView
            contentInsetAdjustmentBehavior="automatic"
            style={styles.scrollView}>
            <View style={styles.body}>
              <View style={styles.sectionContainer}>
                <Text style={styles.sectionTitle}>Добро пожаловать</Text>
                <Text style={styles.sectionDescription}>
                  в умное голосование для нерезидентов!
                </Text>
              </View>
              {sections}
            </View>
          </ScrollView>
        </SafeAreaView>
      </>
    );
  }
}

const styles = StyleSheet.create({
  scrollView: {
    backgroundColor: Colors.lighter,
  },
  body: {
    backgroundColor: Colors.white,
  },
  sectionContainer: {
    marginTop: 32,
    paddingHorizontal: 24,
  },
  sectionTitle: {
    textAlign: 'center',
    fontSize: 24,
    fontWeight: '600',
    color: Colors.black,
  },
  sectionDescription: {
    textAlign: 'center',
    marginTop: 8,
    fontSize: 18,
    fontWeight: '400',
    color: Colors.dark,
  },
});
